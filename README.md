## what
a simple web app that hits omdb for more info on a list of movies i like

## todo
- [ ] add react component for display of movie list
- [ ] make items markable as 'watched'. 
    - [ ] store in indexeddb (browser storage) or sqllite (server-side storage)?
    - [ ] currently movie list is gathered all at once every time we visit /api/movies. Cache + periodically refresh?
- [ ] use imdb id associated with title from omdb to link to imdb page? 
- [ ] add new movies to watch list
- [ ] remove movies from watch list

## run

```bash
(cd src; go run main.go)
```