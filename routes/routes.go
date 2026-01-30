package routes

import (
	"html/template"

	"github.com/lyuangg/gadmin/app"
	"github.com/lyuangg/gadmin/controllers"
	"github.com/lyuangg/gadmin/middleware"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

// devTemplateRenderer 开发模式下每次请求重新加载模板，实现热重载
type devTemplateRenderer struct {
	app *app.App
}

func (r *devTemplateRenderer) Instance(name string, data interface{}) render.Render {
	renderer := createRenderer()
	return renderer.Instance(name, data)
}

func createTemplate(name string, files ...string) *template.Template {
	tmpl, err := template.New(name).Delims("[[", "]]").ParseFiles(files...)
	if err != nil {
		panic("解析模板失败: " + err.Error())
	}
	return tmpl
}

func createRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	loginTmpl := createTemplate("auth/login.html", "templates/auth/login.html")
	r.Add("auth/login.html", loginTmpl)

	adminPages := []string{
		"admin/index.html",
		"admin/users.html",
		"admin/roles.html",
		"admin/permissions.html",
		"admin/dictionaries.html",
		"admin/operation_logs.html",
		"admin/password.html",
		"admin/avatar.html",
	}

	for _, page := range adminPages {
		pageTmpl := createTemplate(page,
			"templates/layouts/admin.html",
			"templates/components/pagination.html",
			"templates/"+page,
		)
		r.Add(page, pageTmpl)
	}
	return r
}

func SetupRoutes(router *gin.Engine, a *app.App) {
	isDevMode := a.Config.GinMode == "debug"

	authController := controllers.NewAuthController(a)
	userController := controllers.NewUserController(a)
	roleController := controllers.NewRoleController(a)
	permissionController := controllers.NewPermissionController(a)
	dictionaryController := controllers.NewDictionaryController(a)
	operationLogController := controllers.NewOperationLogController(a)

	if isDevMode {
		router.HTMLRender = &devTemplateRenderer{app: a}
	} else {
		router.HTMLRender = createRenderer()
	}

	router.Use(middleware.TraceIDMiddleware())
	router.Use(middleware.LoggingMiddleware(a))
	router.Static("/static", "./static")

	router.GET("/login", func(c *gin.Context) {
		c.HTML(200, "auth/login.html", nil)
	})

	api := router.Group("/api")
	api.Use(middleware.RecoveryMiddleware(a))
	{
		api.POST("/login", authController.Login)
		api.GET("/captcha", authController.GetCaptcha)
	}

	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware(a))
	{
		admin.GET("", func(c *gin.Context) {
			c.HTML(200, "admin/index.html", gin.H{
				"PageTitle": "后台管理系统",
			})
		})
		admin.GET("/users", func(c *gin.Context) {
			c.HTML(200, "admin/users.html", gin.H{
				"PageTitle": "用户管理 - 后台管理系统",
			})
		})
		admin.GET("/roles", func(c *gin.Context) {
			c.HTML(200, "admin/roles.html", gin.H{
				"PageTitle": "角色管理 - 后台管理系统",
			})
		})
		admin.GET("/permissions", func(c *gin.Context) {
			c.HTML(200, "admin/permissions.html", gin.H{
				"PageTitle": "权限管理 - 后台管理系统",
			})
		})
		admin.GET("/dictionaries", func(c *gin.Context) {
			c.HTML(200, "admin/dictionaries.html", gin.H{
				"PageTitle": "字典管理 - 后台管理系统",
			})
		})
		admin.GET("/operation-logs", func(c *gin.Context) {
			c.HTML(200, "admin/operation_logs.html", gin.H{
				"PageTitle": "操作日志 - 后台管理系统",
			})
		})
		admin.GET("/password", func(c *gin.Context) {
			c.HTML(200, "admin/password.html", gin.H{
				"PageTitle": "修改密码 - 后台管理系统",
			})
		})
		admin.GET("/avatar", func(c *gin.Context) {
			c.HTML(200, "admin/avatar.html", gin.H{
				"PageTitle": "更换头像 - 后台管理系统",
			})
		})

		adminAPI := admin.Group("/api")
		adminAPI.Use(middleware.RecoveryMiddleware(a))
		adminAPI.Use(middleware.OperationLogMiddleware(a))
		{
			adminAPI.POST("/logout", authController.Logout)
			adminAPI.PUT("/profile/password", authController.ChangePassword)
			adminAPI.PUT("/profile/avatar", authController.UpdateAvatar)
			adminAPI.GET("/user/permissions", authController.GetUserPermissions)

			adminAPIWithPermission := adminAPI.Group("").Use(middleware.PermissionMiddleware(a))
			{
				RegisterRouteWithPermission(adminAPIWithPermission, "GET", "/users", "查询用户列表", "用户管理", userController.GetUsers)
				RegisterRouteWithPermission(adminAPIWithPermission, "POST", "/users", "创建用户", "用户管理", userController.CreateUser)
				RegisterRouteWithPermission(adminAPIWithPermission, "PUT", "/users/:id", "更新用户", "用户管理", userController.UpdateUser)
				RegisterRouteWithPermission(adminAPIWithPermission, "DELETE", "/users/:id", "删除用户", "用户管理", userController.DeleteUser)
				RegisterRouteWithPermission(adminAPIWithPermission, "POST", "/users/:id/reset-password", "重置用户密码", "用户管理", userController.ResetPassword)
				RegisterRouteWithPermission(adminAPIWithPermission, "PUT", "/users/:id/toggle-status", "切换用户状态", "用户管理", userController.ToggleStatus)

				RegisterRouteWithPermission(adminAPIWithPermission, "GET", "/roles", "查询角色列表", "角色管理", roleController.GetRoles)
				RegisterRouteWithPermission(adminAPIWithPermission, "POST", "/roles", "创建角色", "角色管理", roleController.CreateRole)
				RegisterRouteWithPermission(adminAPIWithPermission, "PUT", "/roles/:id", "更新角色", "角色管理", roleController.UpdateRole)
				RegisterRouteWithPermission(adminAPIWithPermission, "DELETE", "/roles/:id", "删除角色", "角色管理", roleController.DeleteRole)
				RegisterRouteWithPermission(adminAPIWithPermission, "PUT", "/roles/:id/permissions", "分配角色权限", "角色管理", roleController.AssignPermissions)

				RegisterRouteWithPermission(adminAPIWithPermission, "GET", "/permissions", "查询权限列表", "权限管理", permissionController.GetPermissions)
				RegisterRouteWithPermission(adminAPIWithPermission, "POST", "/permissions", "创建权限", "权限管理", permissionController.CreatePermission)
				RegisterRouteWithPermission(adminAPIWithPermission, "PUT", "/permissions/:id", "更新权限", "权限管理", permissionController.UpdatePermission)
				RegisterRouteWithPermission(adminAPIWithPermission, "DELETE", "/permissions/:id", "删除权限", "权限管理", permissionController.DeletePermission)
				RegisterRouteWithPermission(adminAPIWithPermission, "POST", "/permissions/batch-delete", "批量删除权限", "权限管理", permissionController.BatchDeletePermissions)

				RegisterRouteWithPermission(adminAPIWithPermission, "GET", "/dictionaries/types", "查询字典类型列表", "字典管理", dictionaryController.GetTypes)
				RegisterRouteWithPermission(adminAPIWithPermission, "POST", "/dictionaries/types", "创建字典类型", "字典管理", dictionaryController.CreateType)
				RegisterRouteWithPermission(adminAPIWithPermission, "PUT", "/dictionaries/types/:id", "更新字典类型", "字典管理", dictionaryController.UpdateType)
				RegisterRouteWithPermission(adminAPIWithPermission, "DELETE", "/dictionaries/types/:id", "删除字典类型", "字典管理", dictionaryController.DeleteType)
				RegisterRouteWithPermission(adminAPIWithPermission, "GET", "/dictionaries/items", "查询字典项列表", "字典管理", dictionaryController.GetItems)
				RegisterRouteWithPermission(adminAPIWithPermission, "GET", "/dictionaries/items/by-code", "根据编码获取字典项", "字典管理", dictionaryController.GetItemsByCode)
				RegisterRouteWithPermission(adminAPIWithPermission, "POST", "/dictionaries/items", "创建字典项", "字典管理", dictionaryController.CreateItem)
				RegisterRouteWithPermission(adminAPIWithPermission, "PUT", "/dictionaries/items/:id", "更新字典项", "字典管理", dictionaryController.UpdateItem)
				RegisterRouteWithPermission(adminAPIWithPermission, "DELETE", "/dictionaries/items/:id", "删除字典项", "字典管理", dictionaryController.DeleteItem)

				RegisterRouteWithPermission(adminAPIWithPermission, "GET", "/operation-logs", "查询操作日志", "系统日志", operationLogController.GetOperationLogs)
			}
		}
	}

	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/login")
	})
}
