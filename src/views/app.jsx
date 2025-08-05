async function getMovies() {
    const res = (await fetch("http://localhost:3000/api/movies/"))
    const movieList = await res.json(); 
    console.log(movieList)
    return movieList
}

// for some reason this function is no longer getting executed at all. Need to figure that out! 
getMovies()

  
export default function MyApp() {
    return (
      <div>
        <h1>Welcome to my app</h1>
        {/* <MyButton /> */}
      </div>
    );
}

ReactDOM.render(<MyApp />, document.getElementById('app'));
