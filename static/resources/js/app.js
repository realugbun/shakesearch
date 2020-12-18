
const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}`).then((response) => {
      response.json().then((results) => {
        Controller.displayResults(results);
      });
    });
  },

  displayResults: function (results) {
    const resultsBody = document.getElementById("results-body");
    resultsBody.innerHTML = results;
  }
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
