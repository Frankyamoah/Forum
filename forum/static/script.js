const newpostButton = document.getElementById("newpost");
newpostButton.addEventListener("click", function() {
  window.location.href = "newpost.html";
});

document.getElementById('loginForm').addEventListener('submit', async (event) => {
  event.preventDefault();
  
  const formData = new FormData(event.target);
  const username = formData.get('username');
  const password = formData.get('password');
  
  if (username && password) {
    const response = await fetch('/login', {
      method: 'POST',
      body: formData
    });
    
    if (response.status === 303) {
      window.location.href = 'forum/static/dashboard.html';
    } else {
      const message = await response.text();
      alert(message);
    }
  } else {
    alert('Please enter a username and password');
  }
});