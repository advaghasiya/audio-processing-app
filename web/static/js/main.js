document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('loginForm');
    const signupForm = document.getElementById('signupForm');
    const uploadForm = document.getElementById('uploadForm');
    const fileList = document.getElementById('fileList');

    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }

    if (signupForm) {
        signupForm.addEventListener('submit', handleSignup);
    }

    if (uploadForm) {
        uploadForm.addEventListener('submit', handleUpload);
        fetchAudioFiles();
    }

    function handleLogin(e) {
        e.preventDefault();
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;

        fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        })
        .then(response => response.json())
        .then(data => {
            if (data.token) {
                localStorage.setItem('token', data.token);
                window.location.href = `/${data.username}`;
            } else {
                alert('Login failed: ' + data.error);
            }
        })
        .catch(error => console.error('Error:', error));
    }

    function handleSignup(e) {
        e.preventDefault();
        const username = document.getElementById('username').value;
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;
        const confirmPassword = document.getElementById('confirmPassword').value;

        fetch('/api/signup', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, email, password, confirmPassword }),
        })
        .then(response => response.json())
        .then(data => {
            if (data.message) {
                alert('Signup successful. Please login.');
                window.location.href = '/login';
            } else {
                alert('Signup failed: ' + data.error);
            }
        })
        .catch(error => console.error('Error:', error));
    }

    function handleUpload(e) {
        e.preventDefault();
        const formData = new FormData();
        const fileInput = document.getElementById('audioFile');
        formData.append('file', fileInput.files[0]);

        fetch('/api/upload', {
            method: 'POST',
            headers: {
                'Authorization': localStorage.getItem('token'),
            },
            body: formData,
        })
        .then(response => response.json())
        .then(data => {
            if (data.message) {
                alert('File uploaded successfully');
                fetchAudioFiles();
            } else {
                alert('Upload failed: ' + data.error);
            }
        })
        .catch(error => console.error('Error:', error));
    }

    function fetchAudioFiles() {
        fetch('/api/audio-files', {
            headers: {
                'Authorization': localStorage.getItem('token'),
            },
        })
        .then(response => response.json())
        .then(files => {
            fileList.innerHTML = '';
            files.forEach(file => {
                const li = document.createElement('li');
                li.textContent = file.name;
                fileList.appendChild(li);
            });
        })
        .catch(error => console.error('Error:', error));
    }
});