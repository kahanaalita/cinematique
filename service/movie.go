package service

// Интерфейс для репозитория Movie
type StoreMovie interface {
	Create(movie Movie) (int, error)
	GetByID(id int) (Movie, error)
	Update(movie Movie) error
	Delete(id int) error
	GetAll() ([]Movie, error)
}

// Структура Movie
type Movie struct {
	ID          int
	Title       string
	Description string
	ReleaseYear int
	Rating      float64
}

// Сервис Movie
type MovieService struct {
	store StoreMovie
}

func NewMovie(store StoreMovie) *MovieService {
	return &MovieService{store: store}
}

func (s *MovieService) Create(movie Movie) (int, error) { return s.store.Create(movie) }
func (s *MovieService) GetByID(id int) (Movie, error)   { return s.store.GetByID(id) }
func (s *MovieService) Update(movie Movie) error        { return s.store.Update(movie) }
func (s *MovieService) Delete(id int) error             { return s.store.Delete(id) }
func (s *MovieService) GetAll() ([]Movie, error)        { return s.store.GetAll() }
