// Theme Management
(function() {
    'use strict';

    const THEME_KEY = 'theme';
    const DARK_THEME = 'dark';
    const LIGHT_THEME = 'light';

    /**
     * Get the current theme from localStorage or system preference
     */
    function getTheme() {
        const stored = localStorage.getItem(THEME_KEY);
        if (stored) {
            return stored;
        }

        // Check system preference
        if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
            return DARK_THEME;
        }

        return LIGHT_THEME;
    }

    /**
     * Apply theme to the document
     */
    function applyTheme(theme) {
        if (theme === DARK_THEME) {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
        localStorage.setItem(THEME_KEY, theme);
        updateToggleIcon(theme);
    }

    /**
     * Update the theme toggle button icon
     */
    function updateToggleIcon(theme) {
        const toggleButtons = document.querySelectorAll('#theme-toggle, #theme-toggle-mobile');
        if (toggleButtons.length === 0) return;

        const sunIcon = `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M12 5a7 7 0 100 14 7 7 0 000-14z" />
        </svg>`;

        const moonIcon = `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
        </svg>`;

        toggleButtons.forEach(toggle => {
            if (theme === DARK_THEME) {
                if(toggle.id === 'theme-toggle-mobile') {
                    toggle.innerHTML = 'Switch to light mode';
                } else {
                    toggle.innerHTML = sunIcon;
                }
                toggle.setAttribute('aria-label', 'Switch to light mode');
            } else {
                if(toggle.id === 'theme-toggle-mobile') {
                    toggle.innerHTML = 'Switch to dark mode';
                } else {
                    toggle.innerHTML = moonIcon;
                }
                toggle.setAttribute('aria-label', 'Switch to dark mode');
            }
        });
    }

    /**
     * Toggle between light and dark themes
     */
    function toggleTheme() {
        const current = getTheme();
        const next = current === DARK_THEME ? LIGHT_THEME : DARK_THEME;
        applyTheme(next);
    }

    /**
     * Initialize theme on page load
     */
    function initTheme() {
        const theme = getTheme();
        applyTheme(theme);

        // Listen for theme toggle clicks
        const toggleButtons = document.querySelectorAll('#theme-toggle, #theme-toggle-mobile');
        toggleButtons.forEach(button => {
            button.addEventListener('click', toggleTheme);
        });

        // Listen for system theme changes
        if (window.matchMedia) {
            window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
                // Only auto-switch if user hasn't manually set a preference
                if (!localStorage.getItem(THEME_KEY)) {
                    applyTheme(e.matches ? DARK_THEME : LIGHT_THEME);
                }
            });
        }
    }

    // Initialize immediately (before DOM ready) to prevent flash
    initTheme();

    // Re-initialize on DOM ready to ensure toggle button is available
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initTheme);
    } else {
        initTheme();
    }

    // Expose theme functions globally if needed
    window.themeManager = {
        getTheme,
        applyTheme,
        toggleTheme
    };
})();
