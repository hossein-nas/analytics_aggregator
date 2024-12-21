package router

import (
	"github.com/gorilla/mux"
	"github.com/hossein-nas/analytics_aggregator/middlewares"
)

func Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/task", middlewares.GetAllTask).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/task", middlewares.CreateTask).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/task/{id}", middlewares.TaskComplete).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/undoTask/{id}", middlewares.UndoTask).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/deleteTask/{id}", middlewares.DeleteTask).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/deleteAllTask", middlewares.DeleteAllTask).Methods("DELETE", "OPTIONS")

	return router
}
