package controller

// Cinematique объединяет сервисы фильмов и актеров в единый фасад
type Cinematique struct {
	Movie ServiceMovie
	Actor ServiceActor
}

// NewCinematique создаёт новый экземпляр фасада
func NewCinematique(movie ServiceMovie, actor ServiceActor) *Cinematique {
	return &Cinematique{
		Movie: movie,
		Actor: actor,
	}
}
