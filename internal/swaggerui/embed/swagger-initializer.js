window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    url: "./swagger_spec",
    dom_id: '#swagger-ui',
    deepLinking: true,
    withCredentials: true,
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
      req.credentials = 'include';
      return req;
    },
  });

  //</editor-fold>

  const loginModal = document.getElementById('login-modal');
  const closeButton = document.querySelector('.close-button');
  const loginForm = document.getElementById('login-form');
  const loginStatus = document.getElementById('login-status');

  const sessionUrl = window.location.origin + '/auth/session';
  const loginUrl = window.location.origin + '/login';
  const logoutUrl = window.location.origin + '/logout';

  let loginBtn = null;
  let currentSession = null;

  const setLoginStatus = (message, isError) => {
    if (!loginStatus) {
      return;
    }

    loginStatus.textContent = message || '';
    loginStatus.style.color = isError ? '#b42318' : '#475467';
  };

  const closeLoginModal = () => {
    loginModal.style.display = 'none';
    setLoginStatus('', false);
  };

  const openLoginModal = () => {
    loginModal.style.display = 'block';
    setLoginStatus('Sign in to create secure auth cookies for this browser session.', false);
  };

  const clearSwaggerHeaderAuth = () => {
    if (window.ui && window.ui.authActions) {
      window.ui.authActions.logout(['bearer']);
    }
  };

  const fetchSession = async () => {
    const response = await fetch(sessionUrl, {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Accept': 'application/json',
      },
    });

    if (!response.ok) {
      return null;
    }

    return response.json();
  };

  const updateLoginButtonState = () => {
    if (!loginBtn) {
      return;
    }

    if (currentSession) {
      const username = currentSession.userName || currentSession.username || 'Current User';
      loginBtn.textContent = 'Logout ' + username;
      loginBtn.onclick = async () => {
        try {
          const response = await fetch(logoutUrl, {
            method: 'POST',
            credentials: 'include',
          });

          if (!response.ok && response.status !== 204) {
            throw new Error('Logout failed.');
          }

          currentSession = null;
          clearSwaggerHeaderAuth();
          updateLoginButtonState();
          alert('Logged out successfully.');
        } catch (error) {
          console.error('Logout error:', error);
          alert(error.message || 'Logout failed.');
        }
      };
      return;
    }

    loginBtn.textContent = 'Login';
    loginBtn.onclick = openLoginModal;
  };

  const syncAuthState = async () => {
    try {
      currentSession = await fetchSession();
    } catch (error) {
      console.error('Session check error:', error);
      currentSession = null;
    }

    clearSwaggerHeaderAuth();
    updateLoginButtonState();
  };

  closeButton.onclick = function() {
    closeLoginModal();
  }

  window.onclick = function(event) {
    if (event.target == loginModal) {
      closeLoginModal();
    }
  }

  loginForm.onsubmit = async function(e) {
    e.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    setLoginStatus('Signing in...', false);

    try {
      const response = await fetch(loginUrl, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        body: JSON.stringify({ username: username, password: password }),
      });

      if (!response.ok) {
        const text = await response.text();
        throw new Error('Login failed: ' + text);
      }

      currentSession = await fetchSession();
      updateLoginButtonState();
      closeLoginModal();
      window.dispatchEvent(new Event('login-success'));
      alert('Login successful.');
    } catch (error) {
      console.error('Login error:', error);
      setLoginStatus(error.message || 'Login failed.', true);
    }
  };

  // --- DOM manipulation logic to add Login/Logout button ---
  const domCheck = setInterval(() => {
    const authWrapper = document.querySelector('.swagger-ui .auth-wrapper');
    if (authWrapper) {
      clearInterval(domCheck);

      loginBtn = document.createElement('button');
      loginBtn.className = 'btn authorize';
      loginBtn.style.marginRight = '10px';
      authWrapper.prepend(loginBtn);
      syncAuthState();
    }
  }, 200);
};
