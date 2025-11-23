// Motion One Animations for Neon UI
(function() {
    'use strict';

    // Wait for DOM to be ready
    function initAnimations() {
        if (typeof Motion === 'undefined') {
            console.warn('Motion library not loaded yet, retrying...');
            setTimeout(initAnimations, 100);
            return;
        }

        const { animate, stagger, spring, inView } = Motion;

        // Animate navbar on load
        animate(
            '.navbar',
            {
                opacity: [0, 1],
                y: [-20, 0]
            },
            {
                duration: 0.6,
                easing: spring({ stiffness: 100, damping: 15 })
            }
        );

        // Animate nav items with stagger
        animate(
            '.nav-item, .theme-toggle',
            {
                opacity: [0, 1],
                scale: [0.8, 1]
            },
            {
                duration: 0.4,
                delay: stagger(0.1, { start: 0.2 }),
                easing: spring({ stiffness: 200, damping: 20 })
            }
        );

        // Animate hero section
        const heroElements = document.querySelectorAll('.hero h1, .hero .subtitle');
        if (heroElements.length > 0) {
            animate(
                heroElements,
                {
                    opacity: [0, 1],
                    y: [30, 0]
                },
                {
                    duration: 0.8,
                    delay: stagger(0.2, { start: 0.3 }),
                    easing: spring({ stiffness: 100, damping: 15 })
                }
            );
        }

        // Animate cards when they come into view
        const cards = document.querySelectorAll('.card, .rule-card, .stat-card, .section');
        cards.forEach((card, index) => {
            inView(card, () => {
                animate(
                    card,
                    {
                        opacity: [0, 1],
                        y: [50, 0],
                        scale: [0.95, 1]
                    },
                    {
                        duration: 0.6,
                        delay: (index % 3) * 0.1, // Stagger by column
                        easing: spring({ stiffness: 100, damping: 15 })
                    }
                );
            });
        });

        // Animate buttons on hover
        const buttons = document.querySelectorAll('.button, .button-primary, .button-danger, .button-success');
        buttons.forEach(button => {
            button.addEventListener('mouseenter', () => {
                animate(
                    button,
                    { scale: 1.05 },
                    { duration: 0.2, easing: spring({ stiffness: 300, damping: 20 }) }
                );
            });

            button.addEventListener('mouseleave', () => {
                animate(
                    button,
                    { scale: 1 },
                    { duration: 0.2, easing: spring({ stiffness: 300, damping: 20 }) }
                );
            });
        });

        // Pulsing glow effect for primary elements
        const glowElements = document.querySelectorAll('.nav-brand a, .button-primary, .status-badge');
        glowElements.forEach(element => {
            animate(
                element,
                {
                    filter: [
                        'drop-shadow(0 0 5px rgba(0, 212, 255, 0.4))',
                        'drop-shadow(0 0 15px rgba(0, 212, 255, 0.8))',
                        'drop-shadow(0 0 5px rgba(0, 212, 255, 0.4))'
                    ]
                },
                {
                    duration: 2,
                    repeat: Infinity,
                    easing: 'ease-in-out'
                }
            );
        });

        // Animate page header
        const pageHeader = document.querySelector('.page-header');
        if (pageHeader) {
            animate(
                pageHeader,
                {
                    opacity: [0, 1],
                    x: [-30, 0]
                },
                {
                    duration: 0.6,
                    easing: spring({ stiffness: 100, damping: 15 })
                }
            );
        }

        // Animate toolbar
        const toolbar = document.querySelector('.toolbar, .connections-toolbar');
        if (toolbar) {
            animate(
                toolbar.children,
                {
                    opacity: [0, 1],
                    y: [20, 0]
                },
                {
                    duration: 0.4,
                    delay: stagger(0.05),
                    easing: spring({ stiffness: 150, damping: 15 })
                }
            );
        }

        // Animate table rows when they come into view
        const tableRows = document.querySelectorAll('.connection-row, .backup-item, .rule-action-card');
        tableRows.forEach((row, index) => {
            inView(row, () => {
                animate(
                    row,
                    {
                        opacity: [0, 1],
                        x: [-20, 0]
                    },
                    {
                        duration: 0.4,
                        delay: (index % 10) * 0.03, // Stagger rows
                        easing: spring({ stiffness: 200, damping: 20 })
                    }
                );
            });
        });

        // Animate modals when they appear
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                mutation.addedNodes.forEach((node) => {
                    if (node.nodeType === 1 && node.classList?.contains('modal-overlay')) {
                        animate(
                            node,
                            { opacity: [0, 1] },
                            { duration: 0.2 }
                        );

                        const modal = node.querySelector('.modal');
                        if (modal) {
                            animate(
                                modal,
                                {
                                    opacity: [0, 1],
                                    scale: [0.9, 1],
                                    y: [30, 0]
                                },
                                {
                                    duration: 0.3,
                                    easing: spring({ stiffness: 200, damping: 20 })
                                }
                            );
                        }
                    }
                });
            });
        });

        observer.observe(document.body, { childList: true, subtree: true });

        // Animate theme toggle icon change
        const themeToggle = document.getElementById('theme-toggle');
        if (themeToggle) {
            const originalClick = themeToggle.onclick;
            themeToggle.addEventListener('click', () => {
                animate(
                    themeToggle,
                    {
                        rotate: [0, 360],
                        scale: [1, 1.2, 1]
                    },
                    {
                        duration: 0.5,
                        easing: spring({ stiffness: 200, damping: 15 })
                    }
                );
            });
        }

        // Subtle floating animation for cards
        const floatingCards = document.querySelectorAll('.card, .stat-card');
        floatingCards.forEach((card, index) => {
            animate(
                card,
                {
                    y: [0, -5, 0]
                },
                {
                    duration: 3 + (index % 3) * 0.5,
                    repeat: Infinity,
                    easing: 'ease-in-out',
                    delay: index * 0.2
                }
            );
        });

        console.log('🎨 Motion One animations initialized');
    }

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initAnimations);
    } else {
        initAnimations();
    }
})();
