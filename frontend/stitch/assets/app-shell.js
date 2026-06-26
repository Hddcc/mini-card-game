(function () {
    const embedded = window.self !== window.top;
    const labels = {
        home: '\u9996\u9875',
        heroes: '\u82f1\u96c4',
        stages: '\u5173\u5361',
        summon: '\u53ec\u5524',
        activity: '\u6d3b\u52a8'
    };
    const pages = [
        { key: 'home', label: labels.home, icon: 'fort', href: '/static/mini_5/code.html' },
        { key: 'heroes', label: labels.heroes, icon: 'swords', href: '/static/mini_3/code.html' },
        { key: 'stages', label: labels.stages, icon: 'scrollable_header', href: '/static/mini_2/code.html' },
        { key: 'summon', label: labels.summon, icon: 'auto_awesome', href: '/static/mini_4/code.html' },
        { key: 'activity', label: labels.activity, icon: 'celebration', href: '/static/mini_6/code.html' }
    ];
    let activeFrame = null;
    let pendingFrame = null;
    let pendingPath = '';
    let battleLocked = false;

    function currentKey(pathname) {
        if (pathname.includes('/mini_3/')) return 'heroes';
        if (pathname.includes('/mini_2/')) return 'stages';
        if (pathname.includes('/mini_4/')) return 'summon';
        if (pathname.includes('/mini_6/')) return 'activity';
        return 'home';
    }

    function installStyle() {
        if (document.getElementById('app-shell-style')) return;
        const style = document.createElement('style');
        style.id = 'app-shell-style';
        style.textContent = `
            [data-app-sidebar] .material-symbols-outlined{font-size:32px;line-height:1;transition:transform .2s ease}
            [data-app-sidebar] a{min-height:72px}
            [data-app-sidebar] a:hover .material-symbols-outlined{transform:scale(1.08)}
            [data-app-shell]{position:fixed;inset:0;z-index:9999;background:#131313;color:#e5e2e1}
            [data-app-frame]{position:absolute;left:18rem;top:0;width:calc(100% - 18rem);height:100%;border:0;background:#131313}
            [data-app-frame][data-pending-frame]{opacity:0;pointer-events:none}
            [data-app-shell][data-battle-locked="true"] [data-app-sidebar]{width:5.5rem}
            [data-app-shell][data-battle-locked="true"] [data-app-sidebar] a{justify-content:center;pointer-events:none;opacity:.55}
            [data-app-shell][data-battle-locked="true"] [data-app-sidebar] span:last-child{display:none}
            [data-app-shell][data-battle-locked="true"] [data-app-frame]{left:5.5rem;width:calc(100% - 5.5rem)}
            @media (max-width:767px){[data-app-shell] [data-app-sidebar]{display:none}[data-app-frame]{left:0;width:100%}}
        `;
        document.head.appendChild(style);
    }

    function linkClass(isActive) {
        if (isActive) {
            return 'text-primary-fixed border-l-4 border-primary-fixed bg-surface-variant px-6 py-5 flex items-center gap-6 transition-all duration-200';
        }
        return 'text-on-surface-variant px-6 py-5 flex items-center gap-6 hover:bg-surface-container-high hover:text-on-surface transition-all duration-200';
    }

    function sidebarHTML(activeKey) {
        const links = pages.map((page) => {
            const isActive = page.key === activeKey;
            const fill = isActive ? " style=\"font-variation-settings: 'FILL' 1;\"" : '';
            const labelClass = isActive
                ? 'hidden md:inline font-title-md text-lg font-bold'
                : 'hidden md:inline font-title-md text-lg font-semibold';
            return `<a class="${linkClass(isActive)}" href="${page.href}" data-shell-link="${page.key}">
                <span class="material-symbols-outlined"${fill}>${page.icon}</span>
                <span class="${labelClass}">${page.label}</span>
            </a>`;
        }).join('');

        return `<aside class="fixed left-0 top-0 h-full z-40 hidden md:flex flex-col pt-24 bg-surface-container-lowest border-r-2 border-outline-variant w-72 transition-all duration-300" data-app-sidebar>
            <div class="flex flex-col gap-4">${links}</div>
        </aside>`;
    }

    function setActive(key) {
        const sidebar = document.querySelector('[data-app-sidebar]');
        if (!sidebar) return;
        pages.forEach((page) => {
            const link = sidebar.querySelector(`[data-shell-link="${page.key}"]`);
            if (!link) return;
            const isActive = page.key === key;
            link.className = linkClass(isActive);
            const icon = link.querySelector('.material-symbols-outlined');
            const label = link.querySelector('span:last-child');
            if (icon) {
                if (isActive) {
                    icon.setAttribute('style', "font-variation-settings: 'FILL' 1;");
                } else {
                    icon.removeAttribute('style');
                }
            }
            if (label) {
                label.className = isActive
                    ? 'hidden md:inline font-title-md text-lg font-bold'
                    : 'hidden md:inline font-title-md text-lg font-semibold';
            }
        });
    }

    function setBattleLocked(locked) {
        battleLocked = !!locked;
        const shell = document.querySelector('[data-app-shell]');
        if (shell) shell.setAttribute('data-battle-locked', battleLocked ? 'true' : 'false');
    }

    function embeddedCleanup() {
        document.querySelectorAll('body > aside, body > nav.fixed.left-0').forEach((node) => node.remove());
        const main = document.querySelector('body > main');
        if (main) {
            ['md:pl-64', 'md:pl-72', 'md:ml-64', 'md:ml-72', 'ml-20', 'ml-24'].forEach((name) => main.classList.remove(name));
        }
        window.top.postMessage({ type: 'mini-xiyou-route', path: location.pathname }, location.origin);
    }

    function createFrame(path) {
        const frame = document.createElement('iframe');
        frame.setAttribute('data-app-frame', '');
        frame.setAttribute('data-pending-frame', '');
        frame.setAttribute('title', 'Mini \u897f\u6e38');
        frame.src = path;
        return frame;
    }

    function showPendingFrame(path) {
        if (!pendingFrame) return;
        const readyFrame = pendingFrame;
        readyFrame.removeAttribute('data-pending-frame');
        if (activeFrame && activeFrame !== readyFrame) activeFrame.remove();
        activeFrame = readyFrame;
        pendingFrame = null;
        pendingPath = '';
        setBattleLocked(false);
        setActive(currentKey(path));
    }

    function navigateFrame(path, push) {
        if (battleLocked) return;
        const shell = document.querySelector('[data-app-shell]');
        if (!shell) return;
        if (pendingFrame) {
            pendingFrame.remove();
            pendingFrame = null;
        }
        pendingPath = path;
        pendingFrame = createFrame(path);
        shell.appendChild(pendingFrame);
        if (push) history.pushState({ shell: true }, '', path);
        setActive(currentKey(path));
    }

    function buildShell() {
        if (document.querySelector('[data-app-shell]')) return;
        installStyle();
        const active = currentKey(location.pathname);
        const shell = document.createElement('div');
        shell.setAttribute('data-app-shell', '');
        shell.innerHTML = sidebarHTML(active);
        document.body.innerHTML = '';
        document.body.appendChild(shell);
        document.body.className = 'bg-background text-on-surface font-body-md overflow-hidden';
        navigateFrame(`${location.pathname}${location.search}${location.hash}`, false);
        bindShellLinks();
    }

    function bindShellLinks() {
        document.addEventListener('click', (event) => {
            const link = event.target.closest('[data-shell-link]');
            if (!link) return;
            event.preventDefault();
            if (battleLocked) return;
            const target = new URL(link.getAttribute('href'), location.origin);
            if (target.pathname === location.pathname) return;
            navigateFrame(target.pathname, true);
        });
    }

    window.addEventListener('message', (event) => {
        if (event.origin !== location.origin || !event.data) return;
        if (event.data.type === 'mini-xiyou-battle-lock') {
            if (activeFrame && event.source === activeFrame.contentWindow) {
                setBattleLocked(!!event.data.locked);
            }
            return;
        }
        if (event.data.type !== 'mini-xiyou-route') return;
        const path = event.data.path || '';
        if (pendingFrame && event.source === pendingFrame.contentWindow) {
            showPendingFrame(path || pendingPath || location.pathname);
            return;
        }
        if (activeFrame && event.source === activeFrame.contentWindow) {
            setActive(currentKey(path));
            if (path && path !== location.pathname) {
                history.replaceState({ shell: true }, '', path);
            }
        }
    });

    window.addEventListener('popstate', () => {
        if (battleLocked) {
            history.pushState({ shell: true }, '', location.pathname);
            return;
        }
        navigateFrame(location.pathname, false);
    });

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', embedded ? embeddedCleanup : buildShell);
    } else {
        (embedded ? embeddedCleanup : buildShell)();
    }
})();
