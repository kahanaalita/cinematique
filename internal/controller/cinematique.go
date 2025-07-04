package controller

// Cinematique объединяет сервисы фильмов и актёров.
type Cinematique struct {
	Movie ServiceMovie
	Actor ServiceActor
}

// NewCinematique создаёт новый экземпляр Cinematique.
func NewCinematique(movie ServiceMovie, actor ServiceActor) *Cinematique {
	return &Cinematique{
		Movie: movie,
		Actor: actor,
	}
}
