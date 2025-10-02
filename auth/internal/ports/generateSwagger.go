package ports

//go:generate swag fmt -d .
//go:generate swag init -d . --parseDependency -g generateSwagger.go -o ../../api/swagger

//	@title						AUTH API
//	@version					1.0
//	@description				API for AUTH service
//	@BasePath					/
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
