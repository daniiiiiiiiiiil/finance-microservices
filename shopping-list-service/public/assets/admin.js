(() => {
  'use strict';
  const $ = s => document.querySelector(s);
  const $$ = s => [...document.querySelectorAll(s)];
  const safe = v => String(v ?? '—').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  let apiBase = localStorage.getItem('fingoals-api-base') || 'http://localhost:8081';
  let token = sessionStorage.getItem('fingoals-token') || '';
  let adminUser = JSON.parse(sessionStorage.getItem('fingoals-user') || '{}');
  const api = async (path, options={}) => {
    const headers = {'Content-Type':'application/json', ...(options.headers || {})};
    if (token) headers.Authorization = 'Bearer ' + token;
    let response;
    try { response = await fetch(apiBase.replace(/\/$/, '') + path, {...options, headers}); }
    catch { throw new Error('Не удалось подключиться к API Gateway. Проверьте адрес сервера.'); }
    const raw = await response.text(); let data={};
    try { data=raw ? JSON.parse(raw) : {}; } catch { data={message:raw}; }
    if (!response.ok) throw new Error(data.message || data.error || `HTTP ${response.status}`);
    return data;
  };
  function hasAdminRole(user, jwt) {
    if (user?.is_admin === true || user?.isAdmin === true || user?.is_admin === 'true' || user?.isAdmin === 'true') return true;
    const roles = [user?.role,user?.roles,user?.authorities,user?.authority].flat().filter(Boolean).join(' ').toUpperCase();
    if (/(^|[\s,[\]"])(ROLE_ADMIN|ADMIN)(?=$|[\s,\]"])/.test(roles)) return true;
    try {
      const part=jwt.split('.')[1]; if(!part) return false;
      const base64=part.replace(/-/g,'+').replace(/_/g,'/');
      const padded=base64.padEnd(base64.length+(4-base64.length%4)%4,'=');
      const bytes=Uint8Array.from(atob(padded),c=>c.charCodeAt(0));
      const claims=JSON.parse(new TextDecoder().decode(bytes));
      if (claims.is_admin===true || claims.isAdmin===true) return true;
      const claimRoles=[claims.role,claims.roles,claims.authorities,claims.authority].flat().filter(Boolean).join(' ').toUpperCase();
      return /(^|[\s,[\]"])(ROLE_ADMIN|ADMIN)(?=$|[\s,\]"])/.test(claimRoles);
    } catch { return false; }
  }
  function showToast(msg) { const el=$('#adminToast'); if(!el) return; el.textContent=msg; el.classList.add('show'); setTimeout(()=>el.classList.remove('show'),3200); }
  function showView(name) {
    const titles={dashboard:['Обзор системы','Сводка и состояние административного API.'],users:['Пользователи','Управление учётными записями и правами доступа.'],transactions:['Транзакции','Операции, доступные текущей учётной записи.']};
    Object.entries({dashboard:'#adminDashboardView',users:'#adminUsersView',transactions:'#adminTransactionsView'}).forEach(([key,selector])=>$(selector).hidden=key!==name);
    $$('[data-admin-view]').forEach(a=>a.classList.toggle('active',a.dataset.adminView===name));
    $('#adminTitle').textContent=titles[name][0]; $('#adminSubtitle').textContent=titles[name][1];
  }
  function authError(message) { $('#adminAuthError').textContent=message; }
  function enterPanel() {
    if (!token) { $('#adminAuth').hidden=false; $('#adminContent').hidden=true; return false; }
    if (!hasAdminRole(adminUser,token)) { $('#adminAuth').hidden=false; $('#adminContent').hidden=true; authError('В ответе авторизации не найдена роль ADMIN.'); $('#adminIdentity').textContent='Недостаточно прав'; return false; }
    $('#adminAuth').hidden=true; $('#adminContent').hidden=false; $('#adminLogout').hidden=false;
    $('#adminIdentity').textContent=adminUser.email || adminUser.username || 'Роль ADMIN';
    $('#adminConnection').textContent='Администратор авторизован'; return true;
  }
  async function loadDashboard() {
    if (!enterPanel()) return;
    $('#statApi').textContent='…'; $('#adminStatusPill').textContent='Проверка'; $('#adminStatusPill').className='admin-status-pill';
    try {
      const data=await api('/api/v1/admin/metrics');
      $('#statUsers').textContent=data.totalUsers ?? data.total_users ?? '—';
      $('#statTransactions').textContent=data.totalTransactions ?? data.total_transactions ?? '—';
      $('#statApi').textContent='Доступен'; $('#statApiNote').textContent=`Общий баланс: ${Number(data.totalBalance ?? data.total_balance ?? 0).toLocaleString('ru-RU',{maximumFractionDigits:2})}`;
      $('#adminStatusPill').textContent='Подключено'; $('#adminStatusPill').className='admin-status-pill ok';
      $('#adminStatusMessage').textContent='Получены реальные метрики из GET /api/v1/admin/metrics.'; $('#adminConnection').textContent='API доступен';
    } catch(e) {
      $('#statApi').textContent='Ошибка'; $('#statApiNote').textContent='Проверьте endpoint';
      $('#adminStatusPill').textContent='Нет данных'; $('#adminStatusPill').className='admin-status-pill bad';
      $('#adminStatusMessage').textContent=`Не удалось получить метрики: ${e.message}. Проверен маршрут GET /api/v1/admin/metrics из admin.proto.`; $('#adminConnection').textContent='Ошибка API';
    }
  }
  async function loadTable(type) {
    if (!enterPanel()) return;
    const isUsers=type==='users';
    const path=isUsers?'/api/v1/admin/users':'/api/v1/finance/transactions?limit=100&offset=0';
    const bodyId=isUsers?'#adminUsersRows':'#adminTransactionsRows';
    $(bodyId).innerHTML=`<tr><td colspan="${isUsers?5:5}" class="empty-state">Загрузка с сервера…</td></tr>`;
    try {
      const result=await api(path);
      const list=isUsers?(result.data||result.users||[]):(result.transactions||result.data||[]);
      if (!list.length) { $(bodyId).innerHTML='<tr><td colspan="5" class="empty-state">Сервер не вернул записей.</td></tr>'; return; }
      if (isUsers) {
        $(bodyId).innerHTML=list.map(x=>{
          const isAdmin=x.is_admin===true||x.isAdmin===true;
          const self=Number(x.id)===Number(adminUser.id);
          return `<tr><td>${safe(x.id)}</td><td><b>${safe(x.full_name||x.fullName||x.name||x.username)}</b></td><td>${safe(x.email)}</td><td><span class="admin-role-badge ${isAdmin?'is-admin':''}">${isAdmin?'ADMIN':'USER'}</span></td><td class="admin-actions"><button class="btn btn-outline admin-action-btn" data-role-id="${Number(x.id)}" data-role-next="${!isAdmin}" ${self?'disabled title="Нельзя менять собственную роль"':''}>${isAdmin?'Снять ADMIN':'Назначить ADMIN'}</button><button class="btn btn-outline admin-action-btn danger" data-delete-id="${Number(x.id)}" ${self?'disabled title="Нельзя удалить себя"':''}>Удалить</button></td></tr>`;
        }).join('');
      } else {
        $(bodyId).innerHTML=list.map(x=>`<tr><td>${safe(x.id)}</td><td>${safe(x.type_transaction||x.typeTransaction||x.type)}</td><td>${safe(x.category)}</td><td>${safe(x.amount)}</td><td>${safe(x.created_at||x.createdAt||x.date)}</td></tr>`).join('');
      }
    } catch(e) { $(bodyId).innerHTML=`<tr><td colspan="5" class="empty-state">${safe(e.message)}. Endpoint: ${path}</td></tr>`; }
  }
  async function changeRole(id, isAdmin) {
    if (!confirm(`Подтвердите ${isAdmin?'назначение роли ADMIN':'снятие роли ADMIN'} для пользователя #${id}?`)) return;
    try {
      await api(`/api/v1/admin/users/${id}/role`,{method:'PATCH',body:JSON.stringify({id:Number(id),is_admin:isAdmin})});
      showToast('Роль пользователя обновлена'); await loadTable('users');
    } catch(e) { showToast(`Не удалось изменить роль: ${e.message}`); }
  }
  async function deleteUser(id) {
    if (!confirm(`Удалить пользователя #${id}? Это действие может быть необратимым.`)) return;
    try { await api(`/api/v1/admin/users/${id}`,{method:'DELETE'}); showToast('Пользователь удалён'); await loadTable('users'); }
    catch(e) { showToast(`Не удалось удалить пользователя: ${e.message}`); }
  }
  $('#adminApiBase').value=apiBase;
  $('#adminLoginForm').addEventListener('submit',async e=>{
    e.preventDefault(); apiBase=$('#adminApiBase').value.trim().replace(/\/$/,''); localStorage.setItem('fingoals-api-base',apiBase);
    $('#adminLoginButton').disabled=true; $('#adminLoginButton').textContent='Выполняется вход…'; authError('');
    try {
      const result=await api('/api/v1/auth/login',{method:'POST',body:JSON.stringify({email:$('#adminEmail').value.trim(),password:$('#adminPassword').value})});
      token=result.token||result.accessToken||result.access_token||''; adminUser=result.user||{};
      if(!token) throw new Error('Сервер не вернул токен авторизации.');
      sessionStorage.setItem('fingoals-token',token); sessionStorage.setItem('fingoals-user',JSON.stringify(adminUser));
      if(!hasAdminRole(adminUser,token)) { enterPanel(); throw new Error('У этой учётной записи нет роли ADMIN.'); }
      enterPanel(); await loadDashboard(); showToast('Вход в панель выполнен');
    } catch(err) { authError(err.message); }
    finally { $('#adminLoginButton').disabled=false; $('#adminLoginButton').textContent='Войти в панель'; }
  });
  $$('[data-admin-view]').forEach(el=>el.addEventListener('click',e=>{e.preventDefault();const view=el.dataset.adminView;showView(view);if(view==='dashboard')loadDashboard();if(view==='users'||view==='transactions')loadTable(view);}));
  $$('[data-admin-load]').forEach(el=>el.addEventListener('click',()=>loadTable(el.dataset.adminLoad)));
  document.addEventListener('click',e=>{
    const role=e.target.closest('[data-role-id]'); if(role){changeRole(Number(role.dataset.roleId),role.dataset.roleNext==='true');return;}
    const del=e.target.closest('[data-delete-id]'); if(del) deleteUser(Number(del.dataset.deleteId));
  });
  $('#adminRefresh').addEventListener('click',()=>{const active=$('.admin-sidebar .nav-link.active')?.dataset.adminView||'dashboard';if(active==='dashboard')loadDashboard();else loadTable(active);});
  $('#adminLogout').addEventListener('click',()=>{token='';adminUser={};sessionStorage.removeItem('fingoals-token');sessionStorage.removeItem('fingoals-user');$('#adminContent').hidden=true;$('#adminAuth').hidden=false;$('#adminLogout').hidden=true;$('#adminIdentity').textContent='Требуется вход';authError('');showToast('Вы вышли из панели');});
  if(token&&hasAdminRole(adminUser,token)){enterPanel();loadDashboard();}else if(token){$('#adminAuth').hidden=false;authError('Сохранённая учётная запись не распознана как ADMIN. Войдите под администратором.');}
})();
