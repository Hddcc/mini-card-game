(function () {
    const MAX_STAMINA = 120;
    let countdownTimer = null;
    let nextStaminaSeconds = 0;

    function token() {
        return localStorage.getItem('token');
    }

    function installStyle() {
        if (document.getElementById('asset-bar-style')) return;
        const style = document.createElement('style');
        style.id = 'asset-bar-style';
        style.textContent = `
            .asset-item{position:relative}
            .asset-tooltip{display:none;position:absolute;right:0;top:calc(100% + 8px);z-index:80;white-space:nowrap;border-radius:4px;background:rgba(0,0,0,.86);padding:4px 8px;color:#fff;font-size:11px;line-height:1.2;box-shadow:0 8px 20px rgba(0,0,0,.35)}
            .asset-item:hover .asset-tooltip{display:block}
        `;
        document.head.appendChild(style);
    }

    function renderAssetBar() {
        installStyle();
        const top = document.querySelector('body > header, body > nav.fixed.top-0');
        if (!top || top.querySelector('[data-asset-bar]')) return;
        const brand = top.firstElementChild;
        if (!brand) return;
        while (brand.nextSibling) brand.nextSibling.remove();
        const wrap = document.createElement('div');
        wrap.className = 'flex items-center gap-4';
        wrap.innerHTML = `
            <div class="flex items-center gap-4 bg-surface-container-lowest px-3 py-1 rounded-full border border-outline-variant/30" data-asset-bar>
                <div class="asset-item flex items-center gap-1.5 border-r border-outline-variant/50 pr-4">
                    <span class="material-symbols-outlined text-primary-fixed text-lg">account_balance_wallet</span>
                    <span class="font-stats-num text-stats-num text-primary-fixed" data-asset-gold>0</span>
                    <span class="asset-tooltip">金币</span>
                </div>
                <div class="asset-item flex items-center gap-1.5 border-r border-outline-variant/50 pr-4">
                    <span class="material-symbols-outlined text-quality-ssr text-lg" style="font-variation-settings: 'FILL' 1;">diamond</span>
                    <span class="font-stats-num text-stats-num text-quality-ssr" data-asset-diamond>0</span>
                    <span class="asset-tooltip">灵玉</span>
                </div>
                <div class="asset-item flex items-center gap-1.5">
                    <span class="material-symbols-outlined text-secondary-container text-lg" style="font-variation-settings: 'FILL' 1;">bolt</span>
                    <span class="font-stats-num text-stats-num text-secondary-container" data-asset-stamina>0/${MAX_STAMINA}</span>
                    <span class="asset-tooltip" data-asset-stamina-tip>体力</span>
                </div>
            </div>
        `;
        top.appendChild(wrap);
    }

    function formatTime(seconds) {
        if (!seconds || seconds <= 0) return '已满';
        const m = Math.floor(seconds / 60);
        const s = seconds % 60;
        return `${m}:${String(s).padStart(2, '0')}`;
    }

    function updateCountdown() {
        const tip = document.querySelector('[data-asset-stamina-tip]');
        if (!tip) return;
        tip.textContent = nextStaminaSeconds > 0 ? `体力 / ${formatTime(nextStaminaSeconds)} 后恢复 1 点` : '体力';
        if (nextStaminaSeconds > 0) nextStaminaSeconds -= 1;
    }

    function updateStageStamina(stamina, maxStamina) {
        const text = document.getElementById('stageStaminaText');
        const fill = document.getElementById('stageStaminaFill');
        if (text) text.textContent = `${stamina}/${maxStamina}`;
        if (fill) fill.style.width = `${Math.max(0, Math.min(100, Math.round(stamina / maxStamina * 100)))}%`;
    }

    async function loadAssetBar() {
        renderAssetBar();
        const auth = token();
        if (!auth) return;
        const res = await fetch('/api/v1/player/assets', {
            headers: { 'Authorization': 'Bearer ' + auth, 'Content-Type': 'application/json' }
        });
        if (res.status === 401) return;
        const json = await res.json();
        if (!json || json.code !== 0 || !json.data) return;
        const data = json.data;
        const maxStamina = data.max_stamina || MAX_STAMINA;
        const stamina = data.stamina || 0;
        const gold = document.querySelector('[data-asset-gold]');
        const diamond = document.querySelector('[data-asset-diamond]');
        const staminaEl = document.querySelector('[data-asset-stamina]');
        if (gold) gold.textContent = data.gold || 0;
        if (diamond) diamond.textContent = data.diamond || 0;
        if (staminaEl) staminaEl.textContent = `${stamina}/${maxStamina}`;
        updateStageStamina(stamina, maxStamina);
        nextStaminaSeconds = data.next_stamina_seconds || 0;
        updateCountdown();
        clearInterval(countdownTimer);
        countdownTimer = setInterval(() => {
            updateCountdown();
            if (nextStaminaSeconds === 0) loadAssetBar();
        }, 1000);
    }

    const originalFetch = window.fetch.bind(window);
    window.fetch = async function () {
        const response = await originalFetch.apply(window, arguments);
        const url = String(arguments[0] || '');
        if (url.includes('/api/v1/stage/fight') || url.includes('/api/v1/gacha/draw') || url.includes('/api/v1/tasks/claim')) {
            setTimeout(loadAssetBar, 80);
        }
        return response;
    };

    window.loadAssetBar = loadAssetBar;
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', loadAssetBar);
    } else {
        loadAssetBar();
    }
})();
