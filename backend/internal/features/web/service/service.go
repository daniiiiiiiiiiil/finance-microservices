package service_web

type WebService struct {
	webRepository WebRepository
}

type WebRepository interface {
	GetFile(filePath string) ([]byte, error)
}

func NewWebService(webRepository WebRepository) *WebService {
	return &WebService{
		webRepository: webRepository,
	}
}
