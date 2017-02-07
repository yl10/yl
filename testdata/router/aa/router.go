package router

import "github.com/astaxie/beego"

type Router struct {
	beego.Controller
}

// @Title 空格后面都是title
// @Description 空格后面都是描述
// @Success 200 {object} models.ZDTProduct.ProductList
// @Param   brand_id    query   int false       "brand id"
// @Param   query   query   string  false       "query of search"
// @Param   segment formData   string  false       "segment"
// @Param   sort    path   string  false       "sort option"
// @Param   dir     body   string  false       "direction asc or desc"
// @Param   offset  header   int     false       "offset"
// @Failure 400 no enough input
// @Failure 500 get products common error
// @router /products [get]`
//格式为：@Param   参数名     [参数类型[formData、query、path、body、header，formData]]   [参数值类型] [是否必须]       [参数描述]
func (r *Router) Get() {

}

func (r *Router) Post() {

}

type Controller struct {
}

// @Title 空格后面都是title
// @Description 空格后面都是描述
// @Success 200 {object} models.ZDTProduct.ProductList
// @Param   brand_id    query   int false       "brand id"
// @Param   query   query   string  false       "query of search"
// @Param   segment formData   string  false       "segment"
// @Param   sort    path   string  false       "sort option"
// @Param   dir     body   string  false       "direction asc or desc"
// @Param   offset  header   int     false       "offset"
// @Failure 400 no enough input
// @Failure 500 get products common error
// @router /products [get]`
//格式为：@Param   参数名     [参数类型[formData、query、path、body、header，formData]]   [参数值类型] [是否必须]       [参数描述]
func (c *Controller) Put() {

}

func (c *Controller) Delete() {

}
