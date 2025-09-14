class EmailVerification {
    constructor() {
        this.verificationStatusEl = document.getElementById('verification-status');
        this.successContentEl = document.getElementById('success-content');
        this.errorContentEl = document.getElementById('error-content');
        this.errorMessageEl = document.getElementById('error-message');
        this.userInfoEl = document.getElementById('user-info');
        this.accessTokenEl = document.getElementById('access-token');
        this.copyTokenBtn = document.getElementById('copy-token');
        this.continueBtn = document.getElementById('continue-btn');
        this.retryBtn = document.getElementById('retry-btn');
        
        this.init();
    }
    
    init() {
        // Parse URL parameters
        const urlParams = new URLSearchParams(window.location.search);
        const token = urlParams.get('token');
        const email = urlParams.get('email');
        
        if (!token || !email) {
            this.showError('Invalid verification link. Missing token or email parameters.');
            return;
        }
        
        // Start verification process
        this.verifyToken(token, email);
        
        // Setup event listeners
        this.setupEventListeners();
    }
    
    setupEventListeners() {
        this.copyTokenBtn?.addEventListener('click', () => this.copyTokenToClipboard());
        this.continueBtn?.addEventListener('click', () => this.redirectToDashboard());
        this.retryBtn?.addEventListener('click', () => this.retryVerification());
    }
    
    async verifyToken(token, email) {
        try {
            const response = await fetch(`/api/auth/email/verify?token=${encodeURIComponent(token)}&email=${encodeURIComponent(email)}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                }
            });
            
            const data = await response.json();
            
            if (response.ok) {
                this.showSuccess(data);
            } else {
                this.showError(data.error || 'Verification failed');
            }
        } catch (error) {
            console.error('Verification error:', error);
            this.showError('Network error. Please check your connection and try again.');
        }
    }
    
    showSuccess(data) {
        this.verificationStatusEl.style.display = 'none';
        this.errorContentEl.style.display = 'none';
        this.successContentEl.style.display = 'block';
        
        // Display user information
        if (data.user) {
            this.userInfoEl.innerHTML = `
                <div class="user-details">
                    <p><strong>Email:</strong> ${data.user.email}</p>
                    <p><strong>User ID:</strong> ${data.user.id}</p>
                    <p><strong>Status:</strong> <span class="status ${data.user.status.toLowerCase()}">${data.user.status}</span></p>
                </div>
            `;
        }
        
        // Display access token
        if (data.access_token) {
            this.accessTokenEl.textContent = data.access_token;
            
            // Store token in localStorage for the application
            localStorage.setItem('luna4_access_token', data.access_token);
            localStorage.setItem('luna4_token_type', data.token_type || 'Bearer');
            localStorage.setItem('luna4_expires_in', data.expires_in || 2592000);
            localStorage.setItem('luna4_user', JSON.stringify(data.user));
        }

        // Auto-redirect
        // this.redirectToDashboard();
    }
    
    showError(errorMessage) {
        this.verificationStatusEl.style.display = 'none';
        this.successContentEl.style.display = 'none';
        this.errorContentEl.style.display = 'block';
        this.errorMessageEl.textContent = errorMessage;
    }
    
    async copyTokenToClipboard() {
        const token = this.accessTokenEl.textContent;
        
        try {
            await navigator.clipboard.writeText(token);
            
            // Show temporary feedback
            const originalText = this.copyTokenBtn.textContent;
            this.copyTokenBtn.textContent = '✓';
            this.copyTokenBtn.style.color = '#28a745';
            
            setTimeout(() => {
                this.copyTokenBtn.textContent = originalText;
                this.copyTokenBtn.style.color = '';
            }, 2000);
        } catch (error) {
            console.error('Failed to copy token:', error);
            
            // Fallback: select the token text
            const range = document.createRange();
            range.selectNode(this.accessTokenEl);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);
            
            alert('Token selected. Please copy manually (Ctrl+C).');
        }
    }
    
    redirectToDashboard() {
        // Parse URL parameters to get redirect
        const urlParams = new URLSearchParams(window.location.search);
        const redirect = urlParams.get('redirect');
        let redirectUrl = redirect || '/app/dashboard';
        
        // Get the access token from localStorage
        const token = localStorage.getItem('luna4_access_token');
        
        // Add access token to redirect URL's query parameters
        if (token) {
            const url = new URL(redirectUrl, window.location.origin);
            url.searchParams.set('accesstoken', token);
            redirectUrl = url.toString();
        }
        
        window.location.href = redirectUrl;
    }
    
    retryVerification() {
        // Reset the page state and try again
        const urlParams = new URLSearchParams(window.location.search);
        const token = urlParams.get('token');
        const email = urlParams.get('email');
        
        if (token && email) {
            this.errorContentEl.style.display = 'none';
            this.verificationStatusEl.style.display = 'block';
            this.verifyToken(token, email);
        } else {
            window.location.href = '/app/';
        }
    }
    
    // Utility method to format time remaining
    formatTimeRemaining(seconds) {
        const days = Math.floor(seconds / (24 * 60 * 60));
        const hours = Math.floor((seconds % (24 * 60 * 60)) / (60 * 60));
        const minutes = Math.floor((seconds % (60 * 60)) / 60);
        
        if (days > 0) {
            return `${days} day${days > 1 ? 's' : ''}`;
        } else if (hours > 0) {
            return `${hours} hour${hours > 1 ? 's' : ''}`;
        } else {
            return `${minutes} minute${minutes > 1 ? 's' : ''}`;
        }
    }
}

// Initialize verification when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new EmailVerification();
});

// Handle page visibility change
document.addEventListener('visibilitychange', () => {
    if (!document.hidden) {
        // Page became visible, could refresh verification status if needed
    }
});