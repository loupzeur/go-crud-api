package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/loupzeur/go-crud-api/utils"
)

//Store db in use
var db *gorm.DB

//Add some specific filtering stuff depending on object fields
var DefaultQueryFilteringFunc = func(r *http.Request, req *gorm.DB, key string, value []string) *gorm.DB {
	//key come from Object Columns field so no injection here
	switch key {
	//exemple of other implementation
	//case "product.id_company":
	//	user, ok := utils.GetAuthenticatedToken(r)
	//	req = req.Where(k+" IN (?) or 1=?", val, ok && user.IsAdmin())
	default:
		req = req.Where(key+" IN (?)", value)
	}
	return req
}

//SetDB to initialize the DB for the API
func SetDB(initDB *gorm.DB) {
	db = initDB
}

//GetDB return the actual DB used
func GetDB() *gorm.DB {
	return db
}

//DeleteAssociation to remove association and data
func DeleteAssociation(tx *gorm.DB, parent Validation, dataType Validation, associationName string, key string, fieldName string) {
	dtype := reflect.TypeOf(dataType)
	data := reflect.New(reflect.SliceOf(dtype))

	tx.Model(parent).Association(associationName).Find(data.Interface())
	for k := 0; k < data.Elem().Len(); k++ {
		v := data.Elem().Index(k).Interface()
		r := reflect.Indirect(reflect.ValueOf(v))
		f := r.FieldByName(fieldName)
		if f.Int() > 0 {
			tx.Delete(v, key, f.Int())
		}
	}
}

//CrudRoutesSpecificURL Generate default CRUD route for object with specific url
func CrudRoutesSpecificURL(models Validation,
	freq func(r *http.Request, req *gorm.DB) *gorm.DB, getallrights utils.RightBits,
	getfunc func(r *http.Request, data interface{}) bool, getrights utils.RightBits,
	crefunc func(r *http.Request, data interface{}) bool, crerights utils.RightBits,
	updfunc func(r *http.Request, data interface{}, data2 interface{}) bool, updrights utils.RightBits,
	delfunc func(r *http.Request, data interface{}) bool, delrights utils.RightBits, url string) utils.Routes {
	parentName := strings.Split(url, "/")
	return utils.Routes{
		utils.Route{"GetAll" + parentName[0] + strings.Title(models.TableName()), "GET", "/api/" + url + models.TableName(),
			func(w http.ResponseWriter, r *http.Request) {
				GenericGetQueryAll(w, r, models, freq)
			}, uint32(getallrights)},
		utils.Route{"Get" + parentName[0] + strings.Title(models.TableName()), "GET", "/api/" + url + models.TableName() + "/{id:[0-9]+}",
			func(w http.ResponseWriter, r *http.Request) {
				GenericGet(w, r, models, getfunc)
			}, uint32(getrights)},
		utils.Route{"Create" + parentName[0] + strings.Title(models.TableName()), "POST", "/api/" + url + models.TableName(),
			func(w http.ResponseWriter, r *http.Request) {
				GenericCreate(w, r, models, crefunc)
			}, uint32(crerights)},
		utils.Route{"Update" + parentName[0] + strings.Title(models.TableName()), "PUT", "/api/" + url + models.TableName() + "/{id:[0-9]+}",
			func(w http.ResponseWriter, r *http.Request) {
				GenericUpdate(w, r, models, updfunc)
			}, uint32(updrights)},
		utils.Route{"Delete" + parentName[0] + strings.Title(models.TableName()), "DELETE", "/api/" + url + models.TableName() + "/{id:[0-9]+}",
			func(w http.ResponseWriter, r *http.Request) {
				GenericDelete(w, r, models, delfunc)
			}, uint32(delrights)},
	}
}

//CrudRoutes Generate default CRUD route for object
func CrudRoutes(models Validation,
	freq func(r *http.Request, req *gorm.DB) *gorm.DB, getallrights utils.RightBits,
	getfunc func(r *http.Request, data interface{}) bool, getrights utils.RightBits,
	crefunc func(r *http.Request, data interface{}) bool, crerights utils.RightBits,
	updfunc func(r *http.Request, data interface{}, data2 interface{}) bool, updrights utils.RightBits,
	delfunc func(r *http.Request, data interface{}) bool, delrights utils.RightBits) utils.Routes {
	return CrudRoutesSpecificURL(models,
		freq, getallrights,
		getfunc, getrights,
		crefunc, crerights,
		updfunc, updrights,
		delfunc, delrights, "")
}

//GetAllFromDb return paginated database : offset, page size, order column
func GetAllFromDb(r *http.Request) (int64, int64, string) {
	page, err := utils.ReadInt(r, "page", 1)
	if err != nil || page < 1 {
		return 0, 0, ""
	}
	pagesize, err := utils.ReadInt(r, "pagesize", 20)
	if err != nil || pagesize <= 0 {
		return 0, 0, ""
	}
	offset := (page - 1) * pagesize
	order := r.FormValue("order")
	return offset, pagesize, order
}

//GenericGetQueryAll return all elements with filters
func GenericGetQueryAll(w http.ResponseWriter, r *http.Request, data Validation, freq func(r *http.Request, req *gorm.DB) *gorm.DB) {
	dtype := reflect.TypeOf(data)
	pages := reflect.New(reflect.SliceOf(dtype)).Interface()
	//Limit and Pagination Part
	offset, pagesize, order := GetAllFromDb(r)
	err := error(nil)
	if offset <= 0 && pagesize <= 0 {
		err = errors.New("error with elements size")
	}
	//Ordering Part
	hasOrders := false //avoid sql injection on orders
	for _, v := range data.OrderColumns() {
		val := strings.Split(order, "_")
		orderDirection := val[len(val)-1]
		if len(val) >= 2 && strings.HasPrefix(order, v) && (orderDirection == "asc" || orderDirection == "desc") {
			hasOrders = true
			if strings.Contains(v, "date") || strings.Contains(v, "id_") { // doesn't work on date :/
				order = v + " " + strings.ToUpper(orderDirection)
			} else {
				order = v + "*1," + v + " " + strings.ToUpper(orderDirection)
			}
			break
		}
	}
	if !hasOrders {
		order = ""
	}
	req := data.QueryAllFromRequest(r, GetDB()).Model(data)

	//Get Default Query
	req = freq(r, req)

	if order != "" {
		req = req.Order(order)
	}

	req = GetQuery(r, req, data.FilterColumns())

	//Execution request Part
	count, resp, err := DefaultCountFunc(r, req)
	if err != nil {
		utils.Respond(w, utils.Message(false, "Error while retrieving data "))
		log.Println("reference split error :", err.Error())
		return
	}

	err = req.Offset(offset).Limit(pagesize).Find(pages).Error
	if err != nil {
		utils.Respond(w, utils.Message(false, "Error while retrieving data"))
		return
	}

	resp["data"] = pages
	resp["total_nb_values"] = count
	resp["current_page"] = offset/pagesize + 1
	resp["size_page"] = pagesize
	utils.Respond(w, resp)
}

var DefaultCountFunc = func(r *http.Request, req *gorm.DB) (int, map[string]interface{}, error) {
	resp := utils.Message(true, "data returned")
	count := 0
	//here you can replace by some other count specific stuff (on group by, ...)
	req.Count(&count)
	return count, resp, nil
}

//GetQuery add url query to gormrequest
func GetQuery(r *http.Request, req *gorm.DB, columns map[string]string) *gorm.DB {
	//Additionnal Querying Part
	urlvars := r.URL.Query()
	//Remove useless parameters to avoid iterating over filters for nothing ^^
	delete(urlvars, "page")
	delete(urlvars, "order")
	delete(urlvars, "pagesize")

	if len(urlvars) > 0 {
		for k, v := range columns {
			if val, ok := urlvars[k]; ok {
				switch v {
				case "in":
					DefaultQueryFilteringFunc(r, req, k, val)
				case "stringlike":
					req = req.Where(k+" LIKE ?", "%"+val[0]+"%")
				case "year":
					req = req.Where("YEAR("+k+") = ?", val[0])
				default:
					if k == "distinct" {
						// :/
						req = req.Where("? IN (select DISTINCT ?)", val[0], val[0])
					} else {
						req = req.Where(k+"=?", val[0])
					}
				}
			}
		}
	}
	return req
}

//Controllers Generic Accessors

//GenericGet default controller for get
func GenericGet(w http.ResponseWriter, r *http.Request, data Validation, f func(r *http.Request, data interface{}) bool) {
	tmp := reflect.New(reflect.TypeOf(data).Elem()).Interface().(Validation)
	err := GetFromID(r, tmp)
	if !f(r, tmp) {
		utils.RespondCode(w, utils.Message(false, "Forbidden"), http.StatusForbidden)
		return
	}
	if err != nil {
		utils.RespondCode(w, utils.Message(false, "Not Found"), http.StatusNotFound)
		return
	}
	resp := utils.Message(true, "success")
	resp["data"] = tmp
	utils.Respond(w, resp)
}

func setUserEmitter(r *http.Request, data Validation) {
	if v, ok := data.(Authed); ok {
		t, authed := utils.GetAuthenticatedToken(r)
		if authed {
			v.SetUserEmitter(t.UserId)
		}
	}

}

//GenericCreate create a new object
func GenericCreate(w http.ResponseWriter, r *http.Request, data Validation, f ...func(r *http.Request, data interface{}) bool) {
	tmp := reflect.New(reflect.TypeOf(data).Elem()).Interface().(Validation)

	err := createFromJSONRequest(r, tmp)
	if err != nil {
		utils.RespondCode(w, utils.Message(false, "Error : "+err.Error()), http.StatusNotAcceptable)
		return
	}
	setUserEmitter(r, tmp)
	actions := len(f)
	reason, ok := tmp.Validate()
	if !ok {
		utils.RespondCode(w, reason, http.StatusNotAcceptable)
		return
	}

	if actions > 0 && !f[0](r, tmp) {
		utils.RespondCode(w, utils.Message(false, "Forbidden"), http.StatusForbidden)
		return
	}
	if err = GetDB().Save(tmp).Error; err != nil {
		utils.RespondCode(w, utils.Message(false, "Error saving"), http.StatusInternalServerError)
		return
	}
	if actions == 2 {
		f[1](r, tmp) //notification, ...
	}
	resp := utils.Message(true, "success")
	resp["data"] = tmp
	utils.Respond(w, resp)
}

//GenericUpdate default updater for controller
func GenericUpdate(w http.ResponseWriter, r *http.Request, data Validation, f func(r *http.Request, data interface{}, data2 interface{}) bool) {
	tmp1 := reflect.New(reflect.TypeOf(data).Elem()).Interface().(Validation)
	tmp2 := reflect.New(reflect.TypeOf(data).Elem()).Interface()

	err := updateFromID(r, tmp1, tmp2)
	setUserEmitter(r, tmp1)
	val, ret := tmp1.Validate()
	if !ret {
		utils.RespondCode(w, val, http.StatusNotAcceptable)
		return
	}
	if !f(r, tmp1, tmp2) {
		utils.RespondCode(w, utils.Message(false, "Forbidden"), http.StatusForbidden)
		return
	}
	if _, err := copy(tmp1, tmp2); err != nil {
		utils.RespondCode(w, utils.Message(false, "Data Error"), http.StatusInternalServerError)
		return
	}
	if err != nil {
		utils.RespondCode(w, utils.Message(false, "Not Found"), http.StatusNotFound)
		return
	}
	setUserEmitter(r, tmp1)
	if err = GetDB().Save(tmp1).Error; err != nil {
		utils.RespondCode(w, utils.Message(false, "Error saving"), http.StatusInternalServerError)
		return
	}
	resp := utils.Message(true, "success")
	resp["data"] = tmp1
	utils.Respond(w, resp)
}

//GenericDelete default deleter for controller
func GenericDelete(w http.ResponseWriter, r *http.Request, data Validation, f func(r *http.Request, data interface{}) bool) {
	tmp := reflect.New(reflect.TypeOf(data).Elem()).Interface().(Validation)
	//tmp := reflect.Zero(reflect.SliceOf(reflect.TypeOf(data))).Interface()
	err := deleteFromID(r, tmp)
	setUserEmitter(r, tmp)
	if !f(r, tmp) {
		utils.RespondCode(w, utils.Message(false, "Forbidden"), http.StatusForbidden)
		return
	}
	if err != nil {
		utils.RespondCode(w, utils.Message(false, "Not Found"), http.StatusNotFound)
		return
	}
	if err = GetDB().Delete(tmp).Error; err != nil {
		utils.RespondCode(w, utils.Message(false, "Error saving"), http.StatusInternalServerError)
		return
	}
	utils.Respond(w, utils.Message(true, "Deletion successful"))
}

//Internals

//Generic Functions for CRUD

func createFromJSONRequest(r *http.Request, data interface{}) error {
	if err := utils.ReadJSON(r, data); err != nil {
		return err
	}
	return nil
}

func deleteFromID(r *http.Request, data Validation) error {
	return data.FindFromRequest(r)
}

//GetFromID Return object from Id
func GetFromID(r *http.Request, data Validation, preloads ...string) error {
	return data.FindFromRequest(r)
}

func updateFromID(r *http.Request, data1 Validation, data2 interface{}) error {
	err := data1.FindFromRequest(r)
	if err != nil {
		return err
	}
	if err := utils.ReadJSON(r, data2); err != nil {
		return err
	}
	return nil
}

//some difference / copy stuff
func copy(dst interface{}, src interface{}) ([]map[string]interface{}, error) {
	dstV := reflect.Indirect(reflect.ValueOf(dst))
	srcV := reflect.Indirect(reflect.ValueOf(src))

	if !dstV.CanAddr() {
		return nil, errors.New("copy to value is unaddressable")
	}

	if srcV.Type() != dstV.Type() {
		return nil, errors.New("different types can be copied")
	}

	dif := []map[string]interface{}{}

	obj, isHistoryAble := dst.(HistoryAble)
	tName := map[string]string{}
	if isHistoryAble {
		tName = obj.GetHistoryFields()
	}

	tf := func(r reflect.Value) string {
		switch r.Kind() {
		case reflect.Struct:
			if strings.HasPrefix(r.Type().String(), "null.") {
				ret := r.MethodByName("ValueOrZero").Call([]reflect.Value{})
				tmp := ""
				if len(ret) > 0 && ret[0].IsValid() {
					tmp = fmt.Sprintf("%v", ret[0].Interface())
					switch tmp {
					case "false":
						tmp = "non"
					case "true":
						tmp = "oui"
					}
				}
				return tmp
			} else {
				return "Type inconnu : " + r.Type().String()
			}
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			ret := []string{}
			for i := 0; i < r.Len(); i++ {
				name := r.Index(i).FieldByName("Name")
				if name.IsValid() {
					ret = append(ret, fmt.Sprintf("%+v", name.Interface()))
				} else {
					ret = append(ret, "...")
				}
			}
			return strings.Join(ret, ",")
		case reflect.Ptr:
			if r.Elem().IsValid() {
				name := r.Elem().FieldByName("Name")
				if name.IsValid() {
					return fmt.Sprintf("%+v", name.Interface())
				} else {
					return "..."
				}
			}
		default:
			//log.Printf("Default global %s\n", r.Kind().String())
		}
		return fmt.Sprintf("%+v", r.Interface())
	}

	for i := 0; i < dstV.NumField(); i++ {
		f := srcV.Field(i)
		if !isZeroOfUnderlyingType(f.Interface()) {
			if isHistoryAble {
				eName := srcV.Type().Field(i).Name
				fName, fExist := tName[eName]
				if fExist { //only append some usefull fields
					nV := tf(srcV.Field(i))
					oV := tf(dstV.Field(i))
					if oV != nV {
						dif = append(dif, map[string]interface{}{
							"field":    fName,
							"newValue": nV,
							"oldValue": oV})
					}
				}
			}
			dstV.Field(i).Set(f)
		}
	}

	if isHistoryAble {
		obj.SetHistory(dif)
	}

	return dif, nil
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
