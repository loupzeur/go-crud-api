# Crud API for gorm.io Object with opentracing support

[![Report](https://goreportcard.com/badge/github.com/loupzeur/go-crud-api)](https://goreportcard.com/report/github.com/loupzeur/go-crud-api)
[![Report](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
[![doc](https://camo.githubusercontent.com/d1a67a692a0fa15f86748f98a790a28b2086e50ee6cc85015010745183b26eed/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f676f2e6465762d7265666572656e63652d626c75653f6c6f676f3d676f266c6f676f436f6c6f723d7768697465)](https://pkg.go.dev/github.com/loupzeur/go-crud-api)
![gopherbadger-tag-do-not-edit]()


## Overview

Provide the crud API of gorm object

## Support the validation interface 
```
type Validation interface {
	TableName() string
	Validate() (map[string]interface{}, bool)
	OrderColumns() []string
	FilterColumns() map[string]string
	FindFromRequest(r *http.Request) error
	QueryAllFromRequest(r *http.Request, q *gorm.DB) *gorm.DB
}
```

## Enable the gorm backend

```
db, _ := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
api.SetDB(db)
db.AutoMigrate(&TestObject{})
db.Use(gormopentracing.New()) //if using tracing
```


## Set your route to the object

```
router := mux.NewRouter().StrictSlash(true)
router.Use(middlewares.JaeggerLogger) //use of jaegger logger
router.Use(middlewares.RecoverWrap)
router.Use(middlewares.JwtAuthentication)

middlewares.Routes = utils.Routes{}.Append(api.CrudRoutes(&TestObject{},
    utils.DefaultQueryAll, utils.NoRight, //QueryAll
    api.DefaultRightAccess, utils.NoRight, //Read
    api.DefaultRightAccess, utils.NoRight, //Create
    api.DefaultRightEdit, utils.NoRight, //Edit
    api.DefaultRightAccess, utils.NoRight, //Delete
))

for _, route := range middlewares.Routes {
    router.Methods(route.Method).Path(route.Pattern).Handler(route.HandlerFunc).Name(route.Name)
}
srv := http.Server{
    ReadTimeout:       0,
    WriteTimeout:      600 * time.Second,
    IdleTimeout:       0,
    ReadHeaderTimeout: 0,
    Addr:              ":" + port,
    Handler:           router,
}
srv.ListenAndServe()
```


## full example ([main.go](https://github.com/loupzeur/go-crud-api/blob/master/main.go))
