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
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem(THEME_KEY, theme);
        updateToggleIcon(theme);
    }

    /**
     * Update the theme toggle button icon
     */
    function updateToggleIcon(theme) {
        const toggle = document.getElementById('theme-toggle');
        if (!toggle) return;

        if (theme === DARK_THEME) {
            toggle.innerHTML = '☀️';
            toggle.setAttribute('aria-label', 'Switch to light mode');
        } else {
            toggle.innerHTML = '🌙';
            toggle.setAttribute('aria-label', 'Switch to dark mode');
        }
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
        const toggle = document.getElementById('theme-toggle');
        if (toggle) {
            toggle.addEventListener('click', toggleTheme);
        }

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
