package dto

type CreateActorRequest struct {
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birth_date"`
}

type UpdateActorRequest struct {
	Name      *string `json:"name,omitempty"`
	Gender    *string `json:"gender,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"`
}

type ActorResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birth_date"`
}

type ActorsListResponse struct {
	Actors []ActorResponse `json:"actors"`
}

type CreateMovieRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ReleaseYear int     `json:"release_year"`
	Rating      float64 `json:"rating"`
}

type UpdateMovieRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Rating      *float64 `json:"rating,omitempty"`
}

type MovieResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ReleaseYear int     `json:"release_year"`
	Rating      float64 `json:"rating"`
}

type MoviesListResponse struct {
	Movies []MovieResponse `json:"movies"`
}
