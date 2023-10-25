module github.com/bzz/scholar-alert-digest

go 1.12

require (
	github.com/antchfx/htmlquery v1.2.0
	github.com/antchfx/xpath v1.1.2 // indirect
	github.com/cheggaaa/pb/v3 v3.0.3
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.8.3
	gitlab.com/golang-commonmark/markdown v0.0.0-20191124021542-fffb4bed7d15
	golang.org/x/oauth2 v0.7.0
	google.golang.org/api v0.114.0
	google.golang.org/grpc v1.56.3 // indirect
	gopkg.in/russross/blackfriday.v2 v2.0.0 // indirect
)

replace gopkg.in/russross/blackfriday.v2 => github.com/russross/blackfriday v2.0.0+incompatible
