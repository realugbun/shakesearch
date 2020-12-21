"use strict";
const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}`).then((response) => {
      response.json().then((results) => {
        Controller.displayResults(results.HTML, results.NumResults);
        
      });
    });
  },
  
  displayResults: function (HTML, numResults) {
    const resultsBody = document.getElementById("results-body");
    const resultsHeader = document.getElementById("results-header")
    const resultsContainer = document.querySelector(".results-container");
    resultsBody.innerHTML = HTML;
    resultsHeader.innerHTML = `<h1>${numResults} Results found!</h1>`
    resultsContainer.classList.remove("hidden") 
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);


    
