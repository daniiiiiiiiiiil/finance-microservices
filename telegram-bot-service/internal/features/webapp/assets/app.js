(() => {
    'use strict';

    const tg = window.Telegram ? window.Telegram.WebApp : null;
    const $ = id => document.getElementById(id);
    const base = location.pathname.replace(/\/+$/, '') || '/telegram/webapp';
    const state = { token: '', gateway: '', map: null, planId: Number(new URLSearchParams(location.search).get('plan') || 0) };

    const money = value => new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 }).format(Number(value) || 0);
    const distance = meters => `${((Number(meters) || 0) / 1000).toFixed(1)} км`;
    const duration = seconds => `${Math.round((Number(seconds) || 0) / 60)} мин`;
    const safe = value => String(value ?? '').replace(/[&<>"']/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[char]));

    function setStatus(text, error = false) {
        const element = $('status');
        element.textContent = text;
        element.hidden = !text;
        element.classList.toggle('error', Boolean(error));
    }

    async function createSession() {
        if (!tg || !tg.initData) {
            throw new Error('Откройте эту страницу через Telegram Web App.');
        }
        tg.ready();
        tg.expand();
        const response = await fetch(`${base}/session`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ initData: tg.initData })
        });
        const data = await response.json().catch(() => ({}));
        if (!response.ok) {
            throw new Error(data.error === 'telegram account is not linked'
                ? 'Аккаунт не привязан. Привяжите Telegram в настройках профиля FinGoals.'
                : 'Не удалось авторизоваться в Web App.');
        }
        state.token = data.token;
        state.gateway = String(data.gatewayUrl || '').replace(/\/$/, '');
        const user = data.user || {};
        if (user.firstName) {
            $('title').textContent = `Карта покупок · ${user.firstName}`;
        }
    }

    async function api(path) {
        const response = await fetch(state.gateway + path, {
            headers: { Authorization: `Bearer ${state.token}`, Accept: 'application/json' }
        });
        const data = await response.json().catch(() => ({}));
        if (!response.ok) {
            throw new Error(data.message || data.error || `Ошибка API ${response.status}`);
        }
        return data;
    }

    async function loadPlan() {
        if (state.planId > 0) {
            return api(`/api/v1/ai/plans/${state.planId}`);
        }
        const list = await api('/api/v1/ai/plans?limit=1&offset=0');
        const plans = Array.isArray(list.plans) ? list.plans : [];
        if (plans.length === 0) {
            return null;
        }
        return api(`/api/v1/ai/plans/${plans[0].id}`);
    }

    function initMap() {
        if (state.map) {
            return state.map;
        }
        state.map = L.map('map', { zoomControl: true, attributionControl: true });
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 19,
            attribution: '&copy; OpenStreetMap'
        }).addTo(state.map);
        return state.map;
    }

    function renderSummary(plan) {
        const cost = plan.cost || {};
        $('costItems').textContent = money(cost.itemsTotal);
        $('costTravel').textContent = money(cost.travelCost);
        $('costDelivery').textContent = money(cost.deliveryCost);
        $('costTotal').textContent = money(cost.total);
        $('summary').hidden = false;

        const recommendation = plan.recommendation || {};
        const route = plan.route || {};
        const parts = [];
        if (recommendation.reason) {
            parts.push(`<b>Рекомендация:</b> ${safe(recommendation.reason)}`);
        }
        if (route.distanceMeters) {
            parts.push(`Маршрут: ${distance(route.distanceMeters)} · ${duration(route.durationSeconds)}`);
        }
        $('recommendation').innerHTML = parts.join('<br>');
    }

    function renderMap(plan) {
        const map = initMap();
        const stores = Array.isArray(plan.stores) ? plan.stores : [];
        const group = L.featureGroup().addTo(map);
        const colors = ['#2f6bff', '#16835a', '#c26a1b', '#7a3fb8', '#0f8b9c'];

        stores.forEach((store, index) => {
            const latitude = Number(store.latitude);
            const longitude = Number(store.longitude);
            if (!latitude || !longitude) {
                return;
            }
            const marker = L.circleMarker([latitude, longitude], {
                radius: 9,
                color: '#ffffff',
                weight: 3,
                fillColor: colors[index % colors.length],
                fillOpacity: 1
            }).addTo(group);
            const prices = (store.prices || []).slice(0, 6).map(price =>
                `<div>${safe(price.itemName)} — <b>${money(price.price)}</b>${price.inStock === false ? ' · нет в наличии' : ''}</div>`
            ).join('');
            const delivery = store.delivery && store.delivery.available
                ? `<div>Доставка: ${money(store.delivery.price)}${store.delivery.durationMin ? ` · ${store.delivery.durationMin} мин` : ''}</div>`
                : '';
            marker.bindPopup(
                `<b>${index + 1}. ${safe(store.name)}</b>` +
                `<div>${safe(store.address)}</div>` +
                `<div style="margin-top:6px">${prices}</div>${delivery}`
            );
        });

        const geometry = parseGeometry(plan.route && plan.route.geometryGeojson);
        if (geometry.length > 1) {
            L.polyline(geometry, { color: '#2f6bff', weight: 4, opacity: 0.85 }).addTo(group);
        } else {
            const order = (plan.route && Array.isArray(plan.route.order) && plan.route.order.length)
                ? plan.route.order
                : stores.map((_, index) => index);
            const points = order
                .map(index => stores[index])
                .filter(store => store && store.latitude && store.longitude)
                .map(store => [Number(store.latitude), Number(store.longitude)]);
            if (points.length > 1) {
                L.polyline(points, { color: '#2f6bff', weight: 4, opacity: 0.85, dashArray: '6 8' }).addTo(group);
            }
        }

        if (group.getLayers().length > 0) {
            map.fitBounds(group.getBounds().pad(0.2));
        } else {
            map.setView([55.751244, 37.618423], 11);
        }
    }

    function parseGeometry(raw) {
        if (!raw) {
            return [];
        }
        try {
            const geometry = typeof raw === 'string' ? JSON.parse(raw) : raw;
            if (!geometry || !geometry.coordinates) {
                return [];
            }
            if (geometry.type === 'LineString') {
                return geometry.coordinates.map(point => [Number(point[1]), Number(point[0])]);
            }
            if (geometry.type === 'MultiLineString') {
                return geometry.coordinates.flat().map(point => [Number(point[1]), Number(point[0])]);
            }
        } catch (error) {
            return [];
        }
        return [];
    }

    function renderStores(plan) {
        const stores = Array.isArray(plan.stores) ? plan.stores : [];
        const container = $('stores');
        container.innerHTML = '';
        if (stores.length === 0) {
            container.innerHTML = '<div class="store-card"><p>Магазины не найдены.</p></div>';
            return;
        }
        stores.forEach((store, index) => {
            const card = document.createElement('article');
            card.className = 'store-card';
            const available = (store.prices || []).filter(price => price.inStock !== false).length;
            const delivery = store.delivery && store.delivery.available
                ? `<span class="pill success">Доставка ${money(store.delivery.price)}</span>`
                : '<span class="pill muted">Без доставки</span>';
            const mapsUrl = `https://www.openstreetmap.org/?mlat=${store.latitude}&mlon=${store.longitude}#map=16/${store.latitude}/${store.longitude}`;
            card.innerHTML =
                `<h3>${index + 1}. ${safe(store.name)}</h3>` +
                `<p>${safe(store.address || 'Адрес не указан')}</p>` +
                `<div class="store-meta">` +
                `<span class="pill">${available} из ${(store.prices || []).length} в наличии</span>` +
                delivery +
                (store.latitude && store.longitude ? `<a class="pill" href="${mapsUrl}" target="_blank" rel="noopener">На карте</a>` : '') +
                `</div>`;
            container.appendChild(card);
        });
    }

    function renderItems(plan) {
        const items = Array.isArray(plan.items) ? plan.items : [];
        const container = $('items');
        container.innerHTML = '';
        if (items.length === 0) {
            container.innerHTML = '<div class="item-row"><span>Список пуст</span></div>';
            return;
        }
        items.forEach(item => {
            const row = document.createElement('div');
            row.className = 'item-row';
            row.innerHTML = `<span>${safe(item.name)}</span><small>${item.quantity || 1} ${safe(item.unit || 'шт.')}</small>`;
            container.appendChild(row);
        });
    }

    function render(plan) {
        if (!plan) {
            setStatus('Сохранённых AI-планов нет. Создайте план в боте.');
            return;
        }
        setStatus('');
        renderSummary(plan);
        renderMap(plan);
        renderStores(plan);
        renderItems(plan);
    }

    async function load() {
        try {
            setStatus('Загрузка плана…');
            await createSession();
            const plan = await loadPlan();
            render(plan);
        } catch (error) {
            setStatus(error.message || 'Не удалось загрузить план.', true);
        }
    }

    $('refresh').addEventListener('click', load);
    load();
})();
