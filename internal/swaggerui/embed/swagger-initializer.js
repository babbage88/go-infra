window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    url: "./swagger_spec",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout",
    oauth2RedirectUrl: '/docs/oauth2-callback',
    requestInterceptor: (req) => {
      const token = localStorage.getItem('accessToken');
      if (token && req.headers) {
          req.headers.Authorization = 'Bearer ' + token;
      }
      return req;
    },
  });

  //</editor-fold>

  const loginModal = document.getElementById('login-modal');
  const closeButton = document.querySelector('.close-button');
  const loginForm = document.getElementById('login-form');

  closeButton.onclick = function() {
    loginModal.style.display = 'none';
  }

  window.onclick = function(event) {
    if (event.target == loginModal) {
      loginModal.style.display = 'none';
    }
  }

  loginForm.onsubmit = function(e) {
    e.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    const loginUrl = window.location.origin + '/login';

    fetch(loginUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    })
    .then(response => {
        if (!response.ok) {
            return response.text().then(text => { throw new Error('Login failed: ' + text) });
        }
        return response.json();
    })
    .then(data => {
      if (data.accessToken) {
        localStorage.setItem('accessToken', data.accessToken);
        const schema = { "type": "apiKey", "in": "header", "name": "Authorization" };
        window.ui.authActions.authorize({ bearer: { name: "bearer", schema: schema, value: "Bearer " + data.accessToken } });
        loginModal.style.display = 'none';
        
        window.dispatchEvent(new Event('login-success'));

        alert('Login successful!');
      } else {
        alert('Login failed: accessToken not found in response.');
      }
    })
    .catch(error => {
      console.error('Login error:', error);
      alert(error.message);
    });
  }

  const token = localStorage.getItem('accessToken');
  if (token) {
      const schema = { "type": "apiKey", "in": "header", "name": "Authorization" };
      window.ui.authActions.authorize({ bearer: { name: "bearer", schema: schema, value: "Bearer " + token } });
  }

  // --- DOM manipulation logic to add Login/Logout button ---
  const domCheck = setInterval(() => {
    const authWrapper = document.querySelector('.swagger-ui .auth-wrapper');
    if (authWrapper) {
      clearInterval(domCheck);

      const loginBtn = document.createElement('button');
      loginBtn.className = 'btn authorize';
      loginBtn.style.marginRight = '10px';

      const updateLoginButtonState = () => {
        const token = localStorage.getItem('accessToken');
        if (token) {
          loginBtn.textContent = 'Logout';
          loginBtn.onclick = () => {
            localStorage.removeItem('accessToken');
            window.ui.authActions.logout(['bearer']);
            updateLoginButtonState(); // Rerender the button
            alert('Logged out successfully!');
          };
        } else {
          loginBtn.textContent = 'Login';
          loginBtn.onclick = () => {
            document.getElementById('login-modal').style.display = 'block';
          };
        }
      };

      window.addEventListener('login-success', updateLoginButtonState);
      
      // Also listen to storage events to sync across tabs
      window.addEventListener('storage', (event) => {
          if (event.key === 'accessToken') {
              updateLoginButtonState();
          }
      });

      updateLoginButtonState(); // Set initial state
      authWrapper.prepend(loginBtn);
    }
  }, 200);
};
