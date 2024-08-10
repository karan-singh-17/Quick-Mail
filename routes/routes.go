package routes

import (
	"net/http"

	_ "github.com/karan-singh-17/Quick-Mail/docs"
	"github.com/karan-singh-17/Quick-Mail/handlers"
	"github.com/karan-singh-17/Quick-Mail/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", httpSwagger.WrapHandler)

	mux.HandleFunc("/api/user/register", handlers.RegisterUser)
	mux.HandleFunc("/api/user/verify/", handlers.VerifyUser)
	mux.HandleFunc("/api/user/login", handlers.Login)
	mux.HandleFunc("/api/user/logout", handlers.LogOut)
	mux.HandleFunc("/api/user/verify-login-code", handlers.VerifyLoginCode)
	mux.Handle("/api/user/curr-user", middleware.AuthMiddleware(http.HandlerFunc(handlers.CurrentUser)))

	mux.Handle("/api/group/create-group", middleware.AuthMiddleware(http.HandlerFunc(handlers.CreateGroup)))
	mux.Handle("/api/group/get-groups", middleware.AuthMiddleware(http.HandlerFunc(handlers.GetAllGroups)))
	mux.Handle("/api/group/execute-group", middleware.AuthMiddleware(http.HandlerFunc(handlers.SendMailToGroup)))
	mux.Handle("/api/group/edit-group", middleware.AuthMiddleware(http.HandlerFunc(handlers.EditGroup)))
	mux.Handle("/api/group/delete-group", middleware.AuthMiddleware(http.HandlerFunc(handlers.DeleteGroup)))
}
