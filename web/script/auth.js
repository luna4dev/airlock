class AuthForm {
    constructor() {
        this.form = document.getElementById('auth-form');
        this.emailInput = document.getElementById('email');
        this.submitBtn = document.getElementById('submit-btn');
        this.btnText = this.submitBtn.querySelector('.btn-text');
        this.btnLoader = this.submitBtn.querySelector('.btn-loader');
        this.messageEl = document.getElementById('message');
        this.countdownEl = document.getElementById('countdown');
        this.countdownTimer = document.getElementById('countdown-timer');

        this.countdownInterval = null;

        this.init();
    }

    init() {
        this.form.addEventListener('submit', (e) => this.handleSubmit(e));
        this.emailInput.addEventListener('input', () => this.clearMessage());
    }

    async handleSubmit(e) {
        e.preventDefault();

        const email = this.emailInput.value.trim();
        const urlParams = new URLSearchParams(window.location.search);
        const redirect = urlParams.get("redirect");

        if (!email || !this.isValidEmail(email)) {
            this.showMessage('Please enter a valid email address.', 'error');
            return;
        }

        this.setLoading(true);
        this.clearMessage();

        try {
            const response = await fetch('/api/auth/email', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, redirect })
            });

            const data = await response.json();

            if (response.ok) {
                this.showMessage('Authentication email sent! Please check your inbox and click the link to sign in.', 'success');
                this.hideFormPermanently();
            } else {
                this.handleError(response.status, data);
            }
        } catch (error) {
            console.error('Network error:', error);
            this.showMessage('Network error. Please check your connection and try again.', 'error');
            this.setLoading(false);
        }
    }

    handleError(status, data) {
        if (status === 429) {
            // Too many requests - show countdown
            const retryAfter = data.retry_after_seconds || 180;
            this.showMessage(`Too many requests. Please wait before trying again.`, 'warning');
            this.startCountdown(retryAfter);
        } else if (status === 404) {
            this.showMessage('Email address not found. Please contact your administrator for access.', 'error');
        } else if (status === 400) {
            this.showMessage(data.error || 'Invalid email address.', 'error');
        } else {
            this.showMessage(data.error || 'An error occurred. Please try again.', 'error');
        }
        this.setLoading(false);
    }

    setLoading(loading, keepDisabled = false) {
        if (loading) {
            this.submitBtn.disabled = true;
            this.btnText.style.display = 'none';
            this.btnLoader.style.display = 'flex';
        } else {
            this.submitBtn.disabled = keepDisabled;
            this.btnText.style.display = 'block';
            this.btnLoader.style.display = 'none';
        }
    }

    showMessage(text, type) {
        this.messageEl.textContent = text;
        this.messageEl.className = `message ${type}`;
        this.messageEl.style.display = 'block';

        // Auto-hide success messages after 5 seconds
        if (type === 'success') {
            setTimeout(() => {
                this.clearMessage();
            }, 5000);
        }
    }

    clearMessage() {
        this.messageEl.style.display = 'none';
        this.messageEl.className = 'message';
    }

    hideFormPermanently() {
        this.form.style.display = 'none';
        this.setLoading(false);
    }

    startCountdown(seconds) {
        this.countdownEl.style.display = 'block';
        this.submitBtn.disabled = true;

        let remaining = seconds;
        this.countdownTimer.textContent = remaining;

        this.countdownInterval = setInterval(() => {
            remaining--;
            this.countdownTimer.textContent = remaining;

            if (remaining <= 0) {
                this.stopCountdown();
            }
        }, 1000);
    }

    stopCountdown() {
        if (this.countdownInterval) {
            clearInterval(this.countdownInterval);
            this.countdownInterval = null;
        }
        this.countdownEl.style.display = 'none';
        this.submitBtn.disabled = false;
    }

    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }
}

// Initialize the form when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new AuthForm();
});

// Handle page visibility change to clear any running timers
document.addEventListener('visibilitychange', () => {
    if (document.hidden) {
        // Page is hidden, could pause timers if needed
    } else {
        // Page is visible again
    }
});