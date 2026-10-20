(() => {
    'use strict';
    const $ = (selector, root = document) => root.querySelector(selector);
    const $$ = (selector, root = document) => [...root.querySelectorAll(selector)];
    const money = value => new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 }).format(Number(value) || 0);
    const number = (value, digits = 4) => new Intl.NumberFormat('ru-RU', { maximumFractionDigits: digits }).format(Number(value));
    const safe = value => String(value ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));

    const authGate = $('#authGate'), authForm = $('#authForm');
    let apiBase = 'http://localhost:8081';
    let token = sessionStorage.getItem('fingoals-token') || '';
    let user = JSON.parse(sessionStorage.getItem('fingoals-user') || '{}');
    let transactions = [], dashboard = {}, analyticsData = null, analyticsPeriod = 'month', page = 'overview', typeFilter = 'all', search = '', loading = false;

    let flowChartInstance = null;
    let categoryChartInstance = null;

    const pages = {
        overview: ['Финансовый обзор', 'Сводка по данным вашего аккаунта.'],
        transactions: ['Транзакции', 'Операции, загруженные из финансового сервиса.'],
        goals: ['Мои цели', 'Накопления на крупные покупки и мечты.'],
        analytics: ['Аналитика', 'Диаграммы рассчитываются по реальным операциям.'],
        currency: ['Курсы валют', 'Получайте курсы и конвертируйте валюты через сервер.']
    };

    const toast = $('#toast'); let toastTimer;
    function notify(message) { toast.textContent = message; toast.classList.add('show'); clearTimeout(toastTimer); toastTimer = setTimeout(() => toast.classList.remove('show'), 3500); }

    const api = async (path, options = {}) => {
        const headers = { ...(options.headers || {}) };
        if (!(options.body instanceof FormData)) {
            headers['Content-Type'] = 'application/json';
        }
        if (token) headers.Authorization = `Bearer ${token}`;
        let response;
        try { response = await fetch(apiBase.replace(/\/$/, '') + path, { ...options, headers }); }
        catch { throw new Error('Не удалось подключиться к API Gateway. Проверьте адрес и доступность сервера.'); }
        const raw = await response.text(); let data = {};
        try { data = raw ? JSON.parse(raw) : {}; } catch { data = { message: raw }; }
        if (!response.ok) throw new Error(data.message || data.error || data.detail || `Ошибка HTTP ${response.status}`);
        return data;
    };

    function showAuth() { authGate.hidden = false; document.body.style.overflow = 'hidden'; }
    function hideAuth() { authGate.hidden = true; document.body.style.overflow = ''; }

    function authMode(mode) {
        const register = mode === 'register';
        $('#authTitle').textContent = register ? 'Создайте аккаунт' : 'С возвращением';
        $('#authSubtitle').textContent = register ? 'Регистрация в FinGoals.' : 'Войдите, чтобы загрузить реальные финансовые данные.';
        $('#authNameWrap').hidden = !register; $('#authPhoneWrap').hidden = !register;
        $('#authName').required = register; $('#authSwitchText').textContent = register ? 'Уже есть аккаунт?' : 'Впервые в FinGoals?';
        $('#authSwitch').textContent = register ? 'Войти' : 'Создать аккаунт'; $('#authSubmit').textContent = register ? 'Зарегистрироваться' : 'Войти';
        $('#authPassword').autocomplete = register ? 'new-password' : 'current-password'; $('#authError').textContent = '';
        authForm.dataset.mode = mode;
    }

    async function loadData() {
        loading = true; $('#connectionStatus').textContent = 'Загружаем данные…';
        try {
            const [list, dash] = await Promise.all([
                api('/api/v1/finance/transactions?limit=500&offset=0'),
                api('/api/v1/finance/dashboard')
            ]);
            const raw = Array.isArray(list) ? list : (list.transactions || list.items || list.data || []);
            transactions = raw.map(t => {
                const rawType = String(t.typeTransaction ?? t.type_transaction ?? t.type ?? '').toLowerCase();
                const type = ['income', 'credit', 'in', 'доход'].includes(rawType) ? 'income' : 'expense';
                const rawDate = t.createdAt ?? t.created_at ?? t.date ?? t.transactionDate ?? t.transaction_date ?? '';
                return { id: t.id, type, amount: Number(t.amount) || 0, category: t.category || 'Без категории', date: String(rawDate).slice(0, 10), title: t.title || t.name || t.description || t.category || (type === 'income' ? 'Доход' : 'Расход') };
            });
            dashboard = dash || {};
            const fullName = user.fullName || user.full_name || user.name || user.email || 'Мой аккаунт';
            $('#userName').textContent = fullName; $('#avatar').textContent = (fullName[0] || 'F').toUpperCase();
            $('#connectionStatus').textContent = 'Данные получены с сервера';
            renderAll();
            if (!$('#currencyView').hidden) loadRates().catch(e => ratesError(e.message));
            if (!$('#goalsView').hidden) loadGoals().catch(e => notify(e.message));
        } catch (error) {
            $('#connectionStatus').textContent = 'Ошибка подключения';
            showAuth();
            if (token) $('#authError').textContent = `Не удалось загрузить данные: ${error.message}`;
            throw error;
        } finally { loading = false; }
    }

    authForm.addEventListener('submit', async event => {
        event.preventDefault();
        apiBase = ($('#apiBase').value.trim() || 'http://localhost:8081').replace(/\/$/, '');
        $('#authError').textContent = ''; $('#authSubmit').disabled = true; $('#authSubmit').textContent = 'Подождите…';
        const register = authForm.dataset.mode === 'register';
        const payload = register
            ? { name: $('#authName').value.trim(), full_name: $('#authName').value.trim(), email: $('#authEmail').value.trim(), password: $('#authPassword').value, phone_number: $('#authPhone').value.trim() }
            : { email: $('#authEmail').value.trim(), password: $('#authPassword').value };
        try {
            const result = await api('/api/v1/auth/' + (register ? 'register' : 'login'), { method: 'POST', body: JSON.stringify(payload) });
            token = result.token || result.accessToken || result.access_token;
            if (!token) throw new Error('Сервер не вернул токен авторизации.');
            user = result.user || {}; sessionStorage.setItem('fingoals-token', token); sessionStorage.setItem('fingoals-user', JSON.stringify(user));
            await loadData(); hideAuth(); notify(register ? 'Аккаунт создан' : 'Вход выполнен');
        } catch (error) { $('#authError').textContent = error.message || 'Не удалось выполнить запрос'; }
        finally { $('#authSubmit').disabled = false; $('#authSubmit').textContent = register ? 'Зарегистрироваться' : 'Войти'; }
    });

    $('#authSwitch').addEventListener('click', () => authMode(authForm.dataset.mode === 'register' ? 'login' : 'register'));

    function logout() {
        token = ''; user = {}; transactions = []; dashboard = {}; analyticsData = null;
        if (flowChartInstance) { flowChartInstance.destroy(); flowChartInstance = null; }
        if (categoryChartInstance) { categoryChartInstance.destroy(); categoryChartInstance = null; }
        sessionStorage.removeItem('fingoals-token'); sessionStorage.removeItem('fingoals-user');
        showAuth(); authMode('login'); renderAll();
    }

    $('#userMenu').addEventListener('click', () => { if (token && confirm('Выйти из аккаунта?')) logout(); });
    $('#refreshBtn').addEventListener('click', () => loadData().then(async () => {
        if (page === 'analytics') await loadAnalytics();
        if (page === 'goals') await loadGoals();
        notify('Данные обновлены');
    }).catch(e => notify(e.message)));

    function dateStart(period) {
        const d = new Date();
        if (period === 'week') { const day = (d.getDay() + 6) % 7; d.setDate(d.getDate() - day); }
        else if (period === 'month') d.setDate(1);
        else if (period === 'year') { d.setMonth(0, 1); }
        else return '';
        d.setHours(0, 0, 0, 0); return d.toISOString().slice(0, 10);
    }

    function filteredTransactions(period = $('#period').value) {
        const start = dateStart(period);
        return transactions.filter(t => !start || !t.date || t.date >= start);
    }

    function totals(items) {
        return items.reduce((acc, t) => { if (t.type === 'income') acc.income += t.amount; else acc.expense += t.amount; return acc; }, { income: 0, expense: 0 });
    }

    function balanceValue() {
        const keys = ['totalBalance', 'total_balance', 'balance', 'total'];
        for (const key of keys) if (dashboard[key] !== undefined && dashboard[key] !== null && Number.isFinite(Number(dashboard[key]))) return Number(dashboard[key]);
        return null;
    }

    function setMoney(id, value) { const el = $('#' + id); if (el) el.textContent = value === null ? 'Нет данных' : money(value); }

    function renderAll() {
        const list = filteredTransactions(), sum = totals(list), all = totals(transactions), balance = balanceValue();
        setMoney('balanceAmount', balance); setMoney('incomeAmount', sum.income); setMoney('expenseAmount', sum.expense);
        setMoney('txIncome', sum.income); setMoney('txExpense', sum.expense); setMoney('txDifference', sum.income - sum.expense);
        renderRecent(list); renderTransactionTable(list); renderFlowChart(list, $('#flowChart')); renderCategoryChart(list, $('#categoryChart'));
        if (analyticsData) renderAnalytics(analyticsData); else {
            ['analyticsIncome','analyticsExpense','analyticsDifference','analyticsNet','analyticsSavingsRate'].forEach(id => { const el = $('#' + id); if (el) el.textContent = 'Загрузка…'; });
        }
    }

    function renderRecent(list) {
        const rows = $('#recentRows'); const recent = [...list].sort((a, b) => (b.date || '').localeCompare(a.date || '')).slice(0, 6);
        rows.innerHTML = recent.length ? recent.map(t => `<tr><td><b>${safe(t.title)}</b></td><td>${safe(t.category)}</td><td>${safe(t.date || '—')}</td><td class="${t.type === 'income' ? 'amount-positive' : 'amount-negative'}">${t.type === 'income' ? '+' : '−'} ${money(t.amount)}</td></tr>`).join('') : '<tr><td colspan="4" class="empty-state">Нет операций за выбранный период</td></tr>';
    }

    function renderTransactionTable(list) {
        const filtered = list.filter(t => (typeFilter === 'all' || t.type === typeFilter) && `${t.title} ${t.category} ${t.type}`.toLowerCase().includes(search.toLowerCase())).sort((a, b) => (b.date || '').localeCompare(a.date || ''));
        $('#transactionRows').innerHTML = filtered.length ? filtered.map(t => `<tr><td><b>${t.type === 'income' ? 'Доход' : 'Расход'}</b><small class="table-note">${safe(t.title)}</small></td><td>${safe(t.category)}</td><td>${safe(t.date || '—')}</td><td class="${t.type === 'income' ? 'amount-positive' : 'amount-negative'}">${t.type === 'income' ? '+' : '−'} ${money(t.amount)}</td><td class="transaction-actions"><button class="mini-action" type="button" data-edit-transaction="${safe(t.id)}">Редактировать</button><button class="mini-action danger" type="button" data-delete-transaction="${safe(t.id)}">Удалить</button></td></tr>`).join('') : '<tr><td colspan="5" class="empty-state">Операций не найдено</td></tr>';
    }

    function transactionModal(transaction = null) {
        if (!token) { notify('Сначала войдите в аккаунт.'); showAuth(); return; }
        const editing = Boolean(transaction);
        const dialog = document.createElement('dialog');
        dialog.className = 'modal transaction-modal';
        dialog.innerHTML = `<form method="dialog"><button class="modal-close" value="cancel" aria-label="Закрыть">×</button><span class="modal-icon">${editing ? '✎' : '+'}</span><h2>${editing ? 'Редактировать операцию' : 'Новая операция'}</h2><p>Изменения будут сохранены в финансовом сервисе.</p><div class="form-fields"><label class="field-label">Тип операции<select class="control" name="type"><option value="expense" ${transaction?.type !== 'income' ? 'selected' : ''}>Расход</option><option value="income" ${transaction?.type === 'income' ? 'selected' : ''}>Доход</option></select></label><label class="field-label">Категория<input class="control" name="category" maxlength="100" required value="${safe(transaction?.category || '')}" placeholder="Например, Продукты"></label><label class="field-label">Сумма, ₽<input class="control" name="amount" type="number" min="0.01" step="0.01" required value="${transaction ? safe(transaction.amount) : ''}" placeholder="0.00"></label></div><p class="transaction-modal-error" role="alert"></p><div class="modal-actions"><button class="btn btn-outline" value="cancel">Отмена</button><button class="btn btn-primary" value="save">${editing ? 'Сохранить' : 'Добавить'}</button></div></form>`;
        document.body.append(dialog);
        dialog.addEventListener('close', () => dialog.remove());
        dialog.querySelectorAll('[value="cancel"]').forEach(button => button.addEventListener('click', event => { event.preventDefault(); dialog.close('cancel'); }));
        dialog.addEventListener('click', event => { if (event.target === dialog) dialog.close('cancel'); });
        dialog.querySelector('form').addEventListener('submit', async event => {
            if (event.submitter?.value !== 'save') { event.preventDefault(); dialog.close('cancel'); return; }
            event.preventDefault();
            const form = event.currentTarget;
            if (!form.reportValidity()) return;
            const data = Object.fromEntries(new FormData(form));
            const payload = { type_transaction: data.type, amount: Number(data.amount), category: data.category.trim() };
            const saveButton = form.querySelector('[value="save"]');
            const errorBox = form.querySelector('.transaction-modal-error');
            saveButton.disabled = true; saveButton.textContent = 'Сохраняем…'; errorBox.textContent = '';
            try {
                if (editing) {
                    await api('/api/v1/finance/transaction/' + encodeURIComponent(transaction.id), { method: 'PUT', body: JSON.stringify({ id: transaction.id, ...payload }) });
                } else {
                    await api('/api/v1/finance/transaction', { method: 'POST', body: JSON.stringify(payload) });
                }
                await loadData(); dialog.close(); notify(editing ? 'Операция обновлена на сервере' : 'Операция добавлена на сервер');
            } catch (error) {
                errorBox.textContent = error.message || 'Не удалось сохранить операцию';
                saveButton.disabled = false; saveButton.textContent = editing ? 'Сохранить' : 'Добавить';
            }
        });
        dialog.showModal();
    }

    async function deleteTransaction(id) {
        const tx = transactions.find(item => String(item.id) === String(id));
        if (!tx || !confirm(`Удалить операцию «${tx.category}» на сумму ${money(tx.amount)}?`)) return;
        try {
            await api('/api/v1/finance/transaction/' + encodeURIComponent(id), { method: 'DELETE' });
            await loadData(); notify('Операция удалена с сервера');
        } catch (error) { notify('Не удалось удалить: ' + (error.message || 'ошибка API')); }
    }

    $('#addTransactionBtn').addEventListener('click', () => transactionModal());
    $('#addTransactionInline').addEventListener('click', () => transactionModal());
    $('#transactionRows').addEventListener('click', event => {
        const edit = event.target.closest('[data-edit-transaction]');
        const remove = event.target.closest('[data-delete-transaction]');
        if (edit) { const tx = transactions.find(item => String(item.id) === String(edit.dataset.editTransaction)); if (tx) transactionModal(tx); }
        if (remove) deleteTransaction(remove.dataset.deleteTransaction);
    });

    function renderFlowChart(list, root) {
        if (!root) return;
        if (!list.length) { root.innerHTML = '<p class="chart-empty">Недостаточно данных для построения диаграммы.</p>'; return; }
        const byDate = {};
        list.forEach(t => { const key = t.date || 'Дата не указана'; if (!byDate[key]) byDate[key] = { income: 0, expense: 0 }; byDate[key][t.type] += t.amount; });
        const rows = Object.entries(byDate).sort((a, b) => a[0].localeCompare(b[0])).slice(-10);
        const max = Math.max(1, ...rows.flatMap(([, v]) => [v.income, v.expense]));
        root.innerHTML = `<div class="flow-legend"><span><i class="legend-square income-square"></i>Доходы</span><span><i class="legend-square expense-square"></i>Расходы</span></div><div class="flow-bars">${rows.map(([date, v]) => `<div class="flow-day" title="${safe(date)}"><div class="flow-pair"><span class="flow-bar income-bar" style="height:${Math.max(2, v.income / max * 150)}px" title="Доход: ${safe(money(v.income))}"></span><span class="flow-bar expense-bar" style="height:${Math.max(2, v.expense / max * 150)}px" title="Расход: ${safe(money(v.expense))}"></span></div><small>${safe(date === 'Дата не указана' ? '—' : date.slice(5))}</small></div>`).join('')}</div><div class="chart-summary"><span>Доходы: <b class="positive-text">${money(totals(list).income)}</b></span><span>Расходы: <b class="negative-text">${money(totals(list).expense)}</b></span></div>`;
    }

    function renderCategoryChart(list, root) {
        if (!root) return;
        const cats = {};
        list.filter(t => t.type === 'expense').forEach(t => { cats[t.category] = (cats[t.category] || 0) + t.amount; });
        const rows = Object.entries(cats).sort((a, b) => b[1] - a[1]); const max = Math.max(1, ...rows.map(x => x[1]));
        root.innerHTML = rows.length ? rows.map(([name, amount]) => `<div class="category-row"><div class="category-row-head"><span>${safe(name)}</span><b>${money(amount)}</b></div><div class="progress-track"><div class="progress-value" style="width:${amount / max * 100}%"></div></div></div>`).join('') : '<p class="chart-empty">Нет расходов за выбранный период.</p>';
    }


    async function loadGoals() {
        const grid = $('#goalsGrid');
        if (!grid) return;
        grid.innerHTML = '<div class="chart-empty">Загрузка целей…</div>';

        try {
            const data = await api('/api/v1/shopping/list');

            let goals = [];
            if (Array.isArray(data)) goals = data;
            else if (Array.isArray(data.list)) goals = data.list;
            else if (Array.isArray(data.shoppings)) goals = data.shoppings;
            else if (data.data && Array.isArray(data.data.list)) goals = data.data.list;

            if (!goals.length) {
                grid.innerHTML = '<div class="chart-empty">У вас пока нет целей. Добавьте первую!</div>';
                return;
            }

            grid.innerHTML = goals.map(goal => {
                const now = Number(goal.amountNow ?? goal.amount_now ?? 0);
                const finish = Number(goal.amountFinish ?? goal.amount_finish ?? 1);
                const percent = Math.min(100, Math.round((now / finish) * 100)) || 0;
                const isDone = goal.completed || percent >= 100;

                let dateStr = 'Без срока';
                const cDate = goal.completionDate || goal.completion_date;
                if (cDate) {
                    dateStr = new Date(cDate).toLocaleDateString('ru-RU');
                }

                const bgImage = goal.imageKey || goal.image_key
                    ? `background-image: url('${apiBase}/api/v1/shopping/image/${goal.id}');`
                    : 'background: linear-gradient(135deg, #e0eaff, #f5f9ff);';

                return `
          <article class="goal-card ${isDone ? 'done' : ''}">
            <div class="goal-image" style="${bgImage}">
              <span class="goal-status ${isDone ? 'done' : ''}">${isDone ? 'Достигнуто' : 'В процессе'}</span>
            </div>
            <div class="goal-content">
              <h3 class="goal-title">${safe(goal.title)}</h3>
              <p class="goal-desc">${safe(goal.description || 'Без описания')}</p>
              <div class="goal-progress-labels">
                <span>${money(now)}</span>
                <span>из ${money(finish)}</span>
              </div>
              <div class="goal-progress-bar">
                <div class="goal-progress-fill" style="width: ${percent}%"></div>
              </div>
              <div class="goal-footer">
                <span>📅 ${dateStr}</span>
                <div style="display: flex; gap: 8px;">
                  <button class="mini-action" onclick="editGoalProgress(${goal.id}, ${now}, ${finish})">Пополнить</button>
                  <button class="mini-action danger" onclick="deleteGoal(${goal.id})">Удалить</button>
                </div>
              </div>
            </div>
          </article>
        `;
            }).join('');

        } catch (error) {
            grid.innerHTML = `<div class="chart-empty">Не удалось загрузить цели: ${safe(error.message)}<br><small>Проверьте маршруты в gateway для shopping-list-service.</small></div>`;
        }
    }

    window.deleteGoal = async (id) => {
        if (!confirm('Удалить эту цель?')) return;
        try {
            await api(`/api/v1/shopping/${id}`, { method: 'DELETE' });
            notify('Цель удалена');
            loadGoals();
        } catch(e) { notify('Ошибка удаления: ' + e.message); }
    };

    // Пополнение цели (Изменение AmountNow через JSON)
    window.editGoalProgress = async (id, currentNow, finish) => {
        const val = prompt(`Сколько вы накопили на эту цель? (Текущая сумма: ${currentNow} ₽)`, currentNow);
        if (val === null) return;

        // Убираем все лишние символы, оставляем только цифры и точку
        const cleanVal = val.replace(',', '.').replace(/[^0-9.]/g, '');
        const newAmount = Number(cleanVal);

        if (isNaN(newAmount) || newAmount < 0) return notify('Введите корректную сумму');

        try {
            // Отправляем как чистый JSON! Бэкенд (gRPC-gateway) отлично его парсит.
            const payload = {
                id: id,
                amount_now: newAmount,
                amount_finish: finish,
                title: 'Обновление цели', // Обязательное поле для валидации
                description: '',
                completed: false
            };

            await api(`/api/v1/shopping/${id}`, {
                method: 'PUT',
                body: JSON.stringify(payload), // Отправляем JSON
                headers: { 'Content-Type': 'application/json' }
            });

            notify('Прогресс цели обновлен');
            loadGoals();
        } catch(e) {
            notify('Ошибка обновления: ' + e.message);
        }
    };

    $('#addGoalBtn').addEventListener('click', () => {
        if (!token) return showAuth();
        const dialog = document.createElement('dialog');
        dialog.className = 'modal transaction-modal';
        dialog.innerHTML = `
      <form method="dialog">
        <button class="modal-close" value="cancel">×</button>
        <span class="modal-icon">🎯</span>
        <h2>Новая цель</h2>
        <p>Опишите вашу мечту и укажите сумму.</p>
        <div class="form-fields">
          <label class="field-label">Название цели<input class="control" name="title" required placeholder="Например, MacBook Pro"></label>
          <label class="field-label">Описание (необязательно)<input class="control" name="description" placeholder="Кратко о цели"></label>
          <label class="field-label">Нужно накопить (₽)<input class="control" name="finish" type="number" min="1" step="0.01" required placeholder="150000"></label>
          <label class="field-label">Уже накоплено (₽)<input class="control" name="now" type="number" min="0" step="0.01" value="0" required placeholder="0"></label>
          <label class="field-label">Срок сдачи<input class="control" name="date" type="date"></label>
          <label class="field-label">Обложка (картинка)<input class="control" name="image" type="file" accept="image/*" style="padding: 8px;"></label>
        </div>
        <p class="transaction-modal-error" role="alert"></p>
        <div class="modal-actions">
          <button class="btn btn-outline" value="cancel">Отмена</button>
          <button class="btn btn-primary" value="save">Создать цель</button>
        </div>
      </form>`;
        document.body.append(dialog);
        dialog.addEventListener('close', () => dialog.remove());
        dialog.querySelectorAll('[value="cancel"]').forEach(b => b.addEventListener('click', e => { e.preventDefault(); dialog.close(); }));

        dialog.querySelector('form').addEventListener('submit', async e => {
            if (e.submitter?.value !== 'save') return;
            e.preventDefault();
            const form = e.currentTarget;
            if (!form.reportValidity()) return;

            const fd = new FormData(form);
            const btn = form.querySelector('[value="save"]');
            const errorBox = form.querySelector('.transaction-modal-error');
            btn.disabled = true; btn.textContent = 'Сохраняем...'; errorBox.textContent = '';

            // 1. Формируем JSON (без файла, чтобы избежать ошибки парсинга)
            const payload = {
                title: fd.get('title'),
                description: fd.get('description'),
                amount_now: Number(fd.get('now')), // Принудительно в число
                amount_finish: Number(fd.get('finish')), // Принудительно в число
            };

            const dateVal = fd.get('date');
            if (dateVal) {
                // Отправляем в формате ISO (YYYY-MM-DDTHH:mm:ssZ), который ждет gRPC
                payload.completion_date = new Date(dateVal).toISOString();
            }

            try {
                // Шаг 1: Отправляем JSON (Картинку сюда не пихаем, чтобы не ломать парсер)
                const createdData = await api('/api/v1/shopping', {
                    method: 'POST',
                    body: JSON.stringify(payload),
                    headers: { 'Content-Type': 'application/json' }
                });

                const newGoalId = createdData.id || createdData.data?.id;

                // Шаг 2: Если прикрепили картинку, загружаем её отдельно
                const file = fd.get('image');
                if (file && file.size > 0 && newGoalId) {
                    const imageData = new FormData();
                    imageData.append('id', newGoalId);
                    imageData.append('image_data', file);
                    imageData.append('filename', file.name);

                    // ВАЖНО: В вашем сервисе (UpdateShopping) роут для обновления обычно PUT /api/v1/shopping/{id}
                    // Отправляем multipart, но только для картинки
                    await api(`/api/v1/shopping/${newGoalId}`, {
                        method: 'PUT',
                        body: imageData
                        // Заголовок Content-Type ставится автоматически браузером для FormData
                    }).catch(imgErr => {
                        console.warn('Цель создана, но картинка не загрузилась:', imgErr);
                        notify('Цель создана, но обложку загрузить не удалось');
                    });
                }

                dialog.close();
                notify('Цель успешно создана!');
                loadGoals();
            } catch(err) {
                errorBox.textContent = err.message;
                btn.disabled = false; btn.textContent = 'Создать цель';
            }
        });
        dialog.showModal();
    });

    // --- АНАЛИТИКА ---
    function unwrapAnalytics(data) {
        let value = data || {};
        for (let i = 0; i < 3; i++) {
            if (value && typeof value === 'object' && !Array.isArray(value)) {
                const nested = value.analytics ?? value.data ?? value.result;
                if (nested && typeof nested === 'object') { value = nested; continue; }
            }
            break;
        }
        return value;
    }

    function analyticsNumber(value) {
        const n = Number(value);
        return Number.isFinite(n) ? n : 0;
    }

    function renderAnalytics(input) {
        const data = unwrapAnalytics(input);
        const income = analyticsNumber(data.income ?? data.total_income);
        const expenses = analyticsNumber(data.expenses ?? data.expense ?? data.total_expenses);
        const net = analyticsNumber(data.net ?? data.difference ?? data.balance ?? (income - expenses));
        const savings = data.savings_rate ?? data.savingsRate;
        setMoney('analyticsIncome', income);
        setMoney('analyticsExpense', expenses);
        setMoney('analyticsDifference', net);
        setMoney('analyticsNet', net);
        const savingsEl = $('#analyticsSavingsRate');
        if (savingsEl) savingsEl.textContent = savings === undefined || savings === null ? 'Нет данных' : `${number(savings, 2)}%`;
        const range = $('#analyticsRange');
        if (range) {
            const from = data.from ?? data.start_date ?? data.startDate;
            const to = data.to ?? data.end_date ?? data.endDate;
            range.textContent = from || to ? `${from ? String(from).slice(0, 10) : '…'} — ${to ? String(to).slice(0, 10) : '…'}` : `Период: ${analyticsPeriodLabel(analyticsPeriod)}`;
        }
        const series = data.series ?? data.points ?? data.timeline ?? [];
        renderAnalyticsSeries(Array.isArray(series) ? series : []);
        const categories = data.top_categories ?? data.topCategories ?? data.categories ?? [];
        renderAnalyticsCategories(Array.isArray(categories) ? categories : []);
        const status = $('#analyticsStatus');
        if (status) status.textContent = 'Аналитика загружена из Finance Service';
    }

    function analyticsPeriodLabel(period) {
        return ({ day: 'день', week: 'неделя', month: 'месяц', year: 'год' })[period] || period;
    }

    function renderAnalyticsSeries(series) {
        const canvas = document.getElementById('analyticsFlowChartCanvas');
        if (!canvas) return;
        const container = canvas.parentElement;
        if (flowChartInstance) { flowChartInstance.destroy(); flowChartInstance = null; }

        const validSeries = series.map(point => {
            const dateStr = point.period ?? point.date ?? point.bucket ?? point.label ?? '';
            let label = '—';
            if (dateStr) {
                const d = new Date(dateStr);
                label = isNaN(d.getTime()) ? dateStr : d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
            }
            return { label, income: Number(point.income ?? point.total_income ?? 0), expense: Number(point.expenses ?? point.expense ?? point.total_expenses ?? 0) };
        });

        if (validSeries.length <= 2) {
            const html = validSeries.map(row => {
                const maxVal = Math.max(row.income, row.expense, 1);
                const incWidth = (row.income / maxVal) * 100;
                const expWidth = (row.expense / maxVal) * 100;
                return `
          <div style="margin-bottom: 30px; padding: 0 10px;">
            <div style="font-weight: 700; font-size: 15px; margin-bottom: 15px; color: #192238;">${row.label}</div>
            <div style="display: flex; align-items: center; margin-bottom: 15px; gap: 15px;">
              <div style="width: 80px; font-size: 13px; color: #6d7a91;">Доходы</div>
              <div style="flex: 1; background: #f0f4f8; height: 24px; border-radius: 6px; overflow: hidden;"><div style="width: ${Math.max(1, incWidth)}%; background: linear-gradient(90deg, #19b985, #67d9b4); height: 100%; border-radius: 6px;"></div></div>
              <div style="width: 130px; text-align: right; font-weight: 700; font-size: 15px; color: #14966b;">${money(row.income)}</div>
            </div>
            <div style="display: flex; align-items: center; gap: 15px;">
              <div style="width: 80px; font-size: 13px; color: #6d7a91;">Расходы</div>
              <div style="flex: 1; background: #f0f4f8; height: 24px; border-radius: 6px; overflow: hidden;"><div style="width: ${Math.max(1, expWidth)}%; background: linear-gradient(90deg, #d64d5c, #f07886); height: 100%; border-radius: 6px;"></div></div>
              <div style="width: 130px; text-align: right; font-weight: 700; font-size: 15px; color: #d64d5c;">${money(row.expense)}</div>
            </div>
          </div>`;
            }).join('');
            container.innerHTML = `<div style="padding: 20px 0;">${html}</div>`;
            return;
        }

        if (!document.getElementById('analyticsFlowChartCanvas')) {
            container.innerHTML = '<canvas id="analyticsFlowChartCanvas"></canvas>';
        }
        const newCanvas = document.getElementById('analyticsFlowChartCanvas');
        const ctx = newCanvas.getContext('2d');
        const labels = validSeries.map(r => r.label);
        const incomeData = validSeries.map(r => r.income);
        const expenseData = validSeries.map(r => r.expense);

        const gradientIncome = ctx.createLinearGradient(0, 0, 0, 300);
        gradientIncome.addColorStop(0, 'rgba(25, 185, 133, 0.4)');
        gradientIncome.addColorStop(1, 'rgba(25, 185, 133, 0.0)');

        flowChartInstance = new Chart(ctx, {
            type: 'bar',
            data: {
                labels,
                datasets: [
                    { type: 'line', label: 'Доходы', data: incomeData, borderColor: '#19b985', backgroundColor: gradientIncome, borderWidth: 3, pointBackgroundColor: '#fff', pointBorderColor: '#19b985', pointBorderWidth: 2, pointRadius: 5, pointHoverRadius: 7, fill: true, tension: 0.4 },
                    { type: 'bar', label: 'Расходы', data: expenseData, backgroundColor: 'rgba(214, 77, 92, 0.7)', borderRadius: 4, barPercentage: 0.5 }
                ]
            },
            options: {
                responsive: true, maintainAspectRatio: false, interaction: { mode: 'index', intersect: false },
                plugins: {
                    legend: { position: 'top', align: 'end', labels: { usePointStyle: true, boxWidth: 8 } },
                    tooltip: { backgroundColor: 'rgba(23, 36, 62, 0.9)', padding: 12, cornerRadius: 8, callbacks: { label: ctx => `${ctx.dataset.label}: ${money(ctx.parsed.y)}` } }
                },
                scales: {
                    x: { grid: { display: false, drawBorder: false }, ticks: { color: '#8b96aa' } },
                    y: { border: { display: false }, grid: { color: '#edf0f6', drawBorder: false, borderDash: [5, 5] }, ticks: { color: '#8b96aa', callback: v => v >= 1000 ? (v / 1000) + 'k' : v }, suggestedMin: 0, suggestedMax: Math.max(...incomeData, ...expenseData) * 1.2 || 100 }
                }
            }
        });
    }

    function renderAnalyticsCategories(categories) {
        const canvas = document.getElementById('analyticsCategoryChartCanvas');
        if (!canvas) return;
        if (categoryChartInstance) categoryChartInstance.destroy();

        let topCats = [...categories].sort((a, b) => Number(b.amount ?? b.total ?? b.expenses ?? b.value ?? 0) - Number(a.amount ?? a.total ?? a.expenses ?? a.value ?? 0));
        const labels = [], data = [], bg = ['#2f79f6', '#19b985', '#f2a93b', '#d64d5c', '#9b59b6', '#8b96aa'];

        if (topCats.length > 5) {
            topCats.slice(0, 5).forEach(cat => { labels.push(cat.category ?? cat.name ?? 'Без категории'); data.push(Number(cat.amount ?? cat.total ?? cat.expenses ?? cat.value ?? 0)); });
            const otherTotal = topCats.slice(5).reduce((s, c) => s + Number(c.amount ?? c.total ?? c.expenses ?? c.value ?? 0), 0);
            if (otherTotal > 0) { labels.push('Другое'); data.push(otherTotal); }
        } else {
            topCats.forEach(cat => { labels.push(cat.category ?? cat.name ?? 'Без категории'); data.push(Number(cat.amount ?? cat.total ?? cat.expenses ?? cat.value ?? 0)); });
        }

        const ctx = canvas.getContext('2d');
        categoryChartInstance = new Chart(ctx, {
            type: 'doughnut',
            data: { labels, datasets: [{ data, backgroundColor: bg, borderWidth: 3, borderColor: '#ffffff', hoverOffset: 6 }] },
            options: {
                responsive: true, maintainAspectRatio: false, cutout: '75%', layout: { padding: 10 },
                plugins: {
                    legend: {
                        position: 'bottom', labels: {
                            usePointStyle: true, boxWidth: 8, padding: 15, font: { size: 12 },
                            generateLabels: chart => chart.data.labels.map((label, i) => ({ text: `${label} (${money(chart.data.datasets[0].data[i])})`, fillStyle: chart.data.datasets[0].backgroundColor[i], index: i }))
                        }
                    },
                    tooltip: { backgroundColor: 'rgba(23, 36, 62, 0.9)', padding: 12, cornerRadius: 8, callbacks: { label: ctx => { const v = ctx.parsed; const t = ctx.dataset.data.reduce((a, b) => a + b, 0); return ` ${money(v)} (${((v / t) * 100).toFixed(1)}%)`; } } }
                }
            }
        });
    }

    async function loadAnalytics() {
        if (!token) { const s = $('#analyticsStatus'); if (s) s.textContent = 'Войдите, чтобы загрузить аналитику.'; return; }
        const status = $('#analyticsStatus');
        if (status) status.textContent = `Загружаем аналитику за период: ${analyticsPeriodLabel(analyticsPeriod)}…`;
        if (flowChartInstance) flowChartInstance.destroy();
        if (categoryChartInstance) categoryChartInstance.destroy();
        try {
            const data = await api('/api/v1/finance/analytics?period=' + encodeURIComponent(analyticsPeriod));
            analyticsData = unwrapAnalytics(data);
            renderAnalytics(analyticsData);
        } catch (error) {
            analyticsData = null;
            if (status) status.textContent = `Не удалось загрузить аналитику с API: ${error.message}`;
            ['analyticsIncome','analyticsExpense','analyticsDifference','analyticsNet','analyticsSavingsRate'].forEach(id => { const el = $('#' + id); if (el) el.textContent = 'Нет данных'; });
            const range = $('#analyticsRange'); if (range) range.textContent = 'Диапазон не получен';
        }
    }

    function setPage(next) {
        if (!pages[next]) next = 'overview'; page = next;
        Object.keys(pages).forEach(key => { const el = $('#' + key + 'View'); if (el) el.hidden = key !== page; });
        $('#pageTitle').textContent = pages[page][0]; $('#pageSubtitle').textContent = pages[page][1];
        $('.breadcrumb').innerHTML = 'ЛИЧНЫЙ КАБИНЕТ <span>/</span> ' + pages[page][0].toLocaleUpperCase('ru-RU');
        $$('[data-page]').forEach(el => el.classList.toggle('active', el.dataset.page === page));
        $('#sidebar').classList.remove('open'); $('#backdrop').classList.remove('visible'); history.replaceState(null, '', '#' + page);
        if (page === 'currency' && token) loadRates().catch(e => ratesError(e.message));
        if (page === 'analytics' && token) loadAnalytics();
        if (page === 'goals' && token) loadGoals();
    }

    $$('[data-page]').forEach(el => el.addEventListener('click', event => { event.preventDefault(); setPage(el.dataset.page); }));
    $('#openSidebar').addEventListener('click', () => { $('#sidebar').classList.add('open'); $('#backdrop').classList.add('visible'); });

    function closeSidebar() { $('#sidebar').classList.remove('open'); $('#backdrop').classList.remove('visible'); }
    $('#closeSidebar').addEventListener('click', closeSidebar); $('#backdrop').addEventListener('click', closeSidebar);

    $('#period').addEventListener('change', renderAll);

    const analyticsPeriodSelect = $('#analyticsPeriod');
    if (analyticsPeriodSelect) analyticsPeriodSelect.addEventListener('change', () => { analyticsPeriod = analyticsPeriodSelect.value; loadAnalytics(); });

    const analyticsRefresh = $('#analyticsRefresh');
    if (analyticsRefresh) analyticsRefresh.addEventListener('click', () => loadAnalytics());

    $('#searchRecords').addEventListener('input', e => { search = e.target.value; renderTransactionTable(filteredTransactions()); });
    $('#typeRecords').addEventListener('change', e => { typeFilter = e.target.value; renderTransactionTable(filteredTransactions()); });

    $('#clearSearch').addEventListener('click', () => {
        search = ''; typeFilter = 'all';
        $('#searchRecords').value = ''; $('#typeRecords').value = 'all';
        renderTransactionTable(filteredTransactions());
    });

    $('#exportBtn').addEventListener('click', () => {
        const list = filteredTransactions(); if (!list.length) { notify('Нет операций для экспорта.'); return; }
        const rows = [['Тип', 'Категория', 'Дата', 'Сумма'], ...list.map(t => [t.type, t.category, t.date, t.amount])];
        const csv = '\ufeff' + rows.map(row => row.map(v => '"' + String(v ?? '').replaceAll('"', '""') + '"').join(';')).join('\r\n');
        const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8;' })); const a = document.createElement('a'); a.href = url; a.download = 'fingoals-transactions.csv'; a.click(); URL.revokeObjectURL(url);
    });

    function ratesArray(data) {
        const root = data.rates ?? data.items ?? data.currencies ?? data.data ?? data;
        if (Array.isArray(root)) return root.map(item => Array.isArray(item) ? [String(item[0]), Number(item[1])] : [String(item.code ?? item.currency ?? item.symbol ?? item.name ?? ''), Number(item.rate ?? item.value ?? item.price ?? item.amount)]).filter(([k, v]) => k && Number.isFinite(v));
        if (root && typeof root === 'object') return Object.entries(root).map(([key, value]) => [key, Number(typeof value === 'object' ? (value.rate ?? value.value ?? value.price) : value)]).filter(([k, v]) => k && Number.isFinite(v));
        return [];
    }

    function ratesError(message) {
        $('#ratesList').innerHTML = `<p class="error-state">Не удалось загрузить курсы: ${safe(message)}</p>`;
        $('#ratesStamp').textContent = 'Проверьте авторизацию, доступность Currency Service и поддерживаемую базовую валюту.';
    }

    async function loadRates() {
        if (!token) { ratesError('Сначала войдите в аккаунт.'); return; }
        const base = $('#ratesBase').value; $('#ratesList').innerHTML = '<p class="chart-empty">Загружаем курсы с сервера…</p>'; $('#ratesStamp').textContent = '';
        const data = await api('/api/v1/currency/' + encodeURIComponent(base)); const entries = ratesArray(data);
        if (!entries.length) { $('#ratesList').innerHTML = '<p class="chart-empty">Сервер ответил, но формат списка курсов не распознан.</p>'; $('#ratesStamp').textContent = 'Ответ API получен. Проверьте формат ответа GET /api/v1/currency/{base}.'; return; }
        $('#ratesList').innerHTML = entries.filter(([code]) => code !== base).map(([code, rate]) => `<div class="rate-row"><span><b>${safe(code)}</b><small>1 ${safe(base)} = ${safe(code)}</small></span><strong>${number(rate, 8)}</strong></div>`).join('') || '<p class="chart-empty">Нет других валют в ответе.</p>';
        $('#ratesStamp').textContent = `Базовая валюта: ${base}. Курсы получены из Currency Service.`;
    }

    $('#loadRates').addEventListener('click', () => loadRates().catch(e => ratesError(e.message)));
    $('#ratesBase').addEventListener('change', () => loadRates().catch(e => ratesError(e.message)));

    $('#currencySwap').addEventListener('click', () => {
        const from = $('#currencyFrom'), to = $('#currencyTo'), old = from.value;
        from.value = to.value; to.value = old;
        $('#currencyResult').textContent = 'Валюты поменяны местами. Нажмите «Конвертировать».';
        $('#currencyStamp').textContent = '';
    });

    $('#currencyConvert').addEventListener('click', async () => {
        const amount = Number($('#currencyAmount').value), from = $('#currencyFrom').value, to = $('#currencyTo').value;
        const result = $('#currencyResult'), stamp = $('#currencyStamp'), button = $('#currencyConvert');
        if (!Number.isFinite(amount) || amount <= 0) { result.textContent = 'Введите сумму больше нуля.'; stamp.textContent = ''; $('#currencyAmount').focus(); return; }
        if (from === to) { result.innerHTML = `Результат конвертации <strong class="converted-value">${safe(number(amount))} ${safe(to)}</strong>`; stamp.textContent = 'Исходная и целевая валюты совпадают.'; return; }
        button.disabled = true; button.textContent = 'Конвертируем…'; result.textContent = 'Запрос к Currency Service…'; stamp.textContent = '';
        try {
            const data = await api('/api/v1/currency/convert', { method: 'POST', body: JSON.stringify({ from, to, amount }) });
            const converted = Number(data.result ?? data.convertedAmount ?? data.converted_amount ?? data.converted ?? data.amountTo);
            if (!Number.isFinite(converted)) throw new Error('В ответе сервера нет распознанной суммы конвертации.');
            result.innerHTML = `Результат конвертации <strong class="converted-value">${safe(number(converted, 6))} ${safe(to)}</strong>`;
            const rate = Number(data.rate ?? data.exchangeRate ?? data.exchange_rate);
            stamp.textContent = Number.isFinite(rate) ? `Курс: 1 ${from} = ${number(rate, 8)} ${to}.` : `Расчёт выполнен Currency Service: ${from} → ${to}.`;
        } catch (error) { result.textContent = error.message || 'Конвертация не выполнена.'; stamp.textContent = 'Проверьте авторизацию и логи currency/gateway.'; }
        finally { button.disabled = false; button.textContent = 'Конвертировать'; }
    });

    const initial = location.hash.slice(1); setPage(pages[initial] ? initial : 'overview');
    if (token) { hideAuth(); loadData().catch(() => {}); } else { showAuth(); authMode('login'); }
})();