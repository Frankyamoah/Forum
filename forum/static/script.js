  let likeButton = document.getElementById("likeButton");
  let likeCount = document.getElementById("likeCount");
  let dislikeButton = document.getElementById("dislikeButton");
  let dislikeCount = document.getElementById("dislikeCount");

  likeButton.onclick = function() {
    likeCount.innerHTML = parseInt(likeCount.innerHTML) + 1;
  };

  dislikeButton.onclick = function() {
    dislikeCount.innerHTML = parseInt(dislikeCount.innerHTML) + 1;
  };

  let postList = document.getElementById("postList");
  let categoryFilter = document.getElementById("category");

  // Extract categories from posts
  let categories = [];
  for (let i = 0; i < postList.children.length; i++) {
    let post = postList.children[i];
    let category = post.querySelector("p.category").textContent;
    if (!categories.includes(category)) {
      categories.push(category);
    }
  }

  // Add categories to the filter section
  categories.forEach(function(category) {
    let option = document.createElement("option");
    option.value = category.toLowerCase();
    option.textContent = category;
    categoryFilter.appendChild(option);
  });

  // Filter posts based on category
  categoryFilter.onchange = function() {
    let selectedCategory = this.value.toLowerCase();
    for (let i = 0; i < postList.children.length; i++) {
      let post = postList.children[i];
      let postCategory = post.querySelector("p.category").textContent.toLowerCase();
      if (selectedCategory === "" || postCategory === selectedCategory) {
        post.style.display = "block";
      } else {
        post.style.display = "none";
      }
    }
  };
