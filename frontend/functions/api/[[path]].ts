// Cloudflare Pages Functions — API backend for food-ordering
// Rewritten from Go/Gin to TypeScript for Workers + D1 + R2

interface Env {
  DB: D1Database;
  UPLOADS: R2Bucket;
  AI_API_KEY?: string;
  AI_BASE_URL?: string;
  AI_MODEL?: string;
  FEISHU_WEBHOOK_URL?: string;
  WECOM_WEBHOOK_URL?: string;
  PUSHPLUS_TOKEN?: string;
  SHOP_NAME?: string;
}

const STOCK_INFINITE = -1;
const VALID_STATUSES: Record<string, string> = { pending: 'pending', paid: 'paid', done: 'done', cancelled: 'cancelled' };

function corsHeaders() {
  return { 'Access-Control-Allow-Origin': '*', 'Access-Control-Allow-Methods': 'GET, POST, PUT, PATCH, DELETE, OPTIONS', 'Access-Control-Allow-Headers': 'Content-Type, Authorization' };
}

function json(data: any, status = 200): Response {
  return new Response(JSON.stringify(data), { status, headers: { 'Content-Type': 'application/json', ...corsHeaders() } });
}

function fail(status: number, msg: string): Response { return json({ error: msg }, status); }
function ok(): Response { return json({ ok: true }); }
function toCamel(o: any): any {
  if (!o || typeof o !== 'object') return o;
  if (Array.isArray(o)) return o.map(toCamel);
  const r: any = {};
  for (const [k, v] of Object.entries(o)) {
    r[k.replace(/_([a-z])/g, (_, l) => l.toUpperCase())] = (v && typeof v === 'object') ? toCamel(v) : v;
  }
  return r;
}
function camelJson(data: any, status = 200): Response { return json(toCamel(data), status); }

function shopName(env: Env): string { return env.SHOP_NAME || '小馆点餐'; }

function statusLabel(s: string): string {
  const m: Record<string, string> = { pending: '待处理', paid: '已支付', done: '已完成', cancelled: '已取消' };
  return m[s] || s;
}

function centerText(s: string, w: number): string {
  const p = Math.max(0, Math.floor((w - s.length) / 2));
  return ' '.repeat(p) + s;
}

function parseRoute(path: string): { resource: string; id: number | null; suffix: string | null } | null {
  const m = path.replace(/^\/api\//, '').match(/^([^/]+)(?:\/(\d+))?(?:\/([^/]+))?/);
  if (!m) return null;
  return { resource: m[1], id: m[2] ? parseInt(m[2], 10) : null, suffix: m[3] || null };
}

// ========== Ingredients ==========
async function listIngredients(env: Env) {
  return camelJson((await env.DB.prepare('SELECT * FROM ingredients ORDER BY id DESC').all()).results);
}
async function createIngredient(req: Request, env: Env) {
  const b = await req.json() as any;
  if (!b.name?.trim()) return fail(400, '请填写名称');
  return camelJson(await env.DB.prepare("INSERT INTO ingredients (name, unit, stock) VALUES (?, ?, ?) RETURNING *").bind(b.name.trim(), b.unit || '', b.stock ?? 0).first());
}
async function updateIngredient(req: Request, env: Env, id: number) {
  if (!await env.DB.prepare('SELECT 1 FROM ingredients WHERE id = ?').bind(id).first()) return fail(404, '原材料不存在');
  const b = await req.json() as any;
  await env.DB.prepare("UPDATE ingredients SET name=?, unit=?, stock=?, updated_at=datetime('now') WHERE id=?").bind(b.name?.trim() ?? '', b.unit ?? '', b.stock ?? 0, id).run();
  return camelJson(await env.DB.prepare('SELECT * FROM ingredients WHERE id = ?').bind(id).first());
}
async function setIngredientStock(req: Request, env: Env, id: number) {
  const existing = await env.DB.prepare('SELECT * FROM ingredients WHERE id = ?').bind(id).first() as any;
  if (!existing) return fail(404, '原材料不存在');
  const b = await req.json() as { stock?: number; infinite?: boolean };
  const stock = b.infinite ? STOCK_INFINITE : (b.stock ?? 0);
  await env.DB.prepare("UPDATE ingredients SET stock=?, updated_at=datetime('now') WHERE id=?").bind(stock, id).run();
  return camelJson(await env.DB.prepare('SELECT * FROM ingredients WHERE id = ?').bind(id).first());
}
async function deleteIngredient(env: Env, id: number) {
  if (!await env.DB.prepare('SELECT 1 FROM ingredients WHERE id = ?').bind(id).first()) return fail(404, '原材料不存在');
  await env.DB.prepare('DELETE FROM ingredients WHERE id = ?').bind(id).run();
  return ok();
}

// ========== Finance ==========
async function listPurchases(req: Request, env: Env) {
  const u = new URL(req.url), b: any[] = [];
  let s = "SELECT p.*, i.name AS ing_name, i.unit AS ing_unit FROM ingredient_purchases p LEFT JOIN ingredients i ON p.ingredient_id = i.id WHERE 1=1";
  if (u.searchParams.get('start')) { s += ' AND p.purchase_date >= ?'; b.push(u.searchParams.get('start')); }
  if (u.searchParams.get('end')) { s += ' AND p.purchase_date <= ?'; b.push(u.searchParams.get('end')); }
  s += ' ORDER BY p.purchase_date DESC, p.id DESC';
  return camelJson((await (b.length ? env.DB.prepare(s).bind(...b) : env.DB.prepare(s)).all()).results);
}
async function createPurchase(req: Request, env: Env) {
  const b = await req.json() as any;
  const ingredientId = Number(b.ingredientId), quantity = Number(b.quantity);
  if (!ingredientId) return fail(400, '请选择原材料');
  if (quantity <= 0) return fail(400, '采购数量必须大于 0');
  const ing = await env.DB.prepare('SELECT * FROM ingredients WHERE id = ?').bind(ingredientId).first() as any;
  if (!ing) return fail(400, '原材料不存在');
  let pd = (b.purchaseDate || '').trim() || new Date().toISOString().slice(0, 10);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(pd)) return fail(400, '采购日期格式应为 YYYY-MM-DD');
  let tc = Number(b.totalCost) || 0, up = Number(b.unitPrice) || 0;
  if (tc <= 0 && up > 0) { tc = up * quantity; } else if (up <= 0 && tc > 0) { up = Math.round(tc / quantity * 100) / 100; } else if (tc > 0 && up > 0) { up = Math.round(tc / quantity * 100) / 100; }
  if (tc <= 0) return fail(400, '采购金额必须大于 0');
  const p = await env.DB.prepare("INSERT INTO ingredient_purchases (ingredient_id, quantity, unit_price, total_cost, purchase_date, note) VALUES (?,?,?,?,?,?) RETURNING *").bind(ingredientId, quantity, Math.round(up*100)/100, Math.round(tc*100)/100, pd, (b.note||'').trim()).first() as any;
  if (ing.stock !== STOCK_INFINITE) await env.DB.prepare("UPDATE ingredients SET stock = stock + ?, updated_at = datetime('now') WHERE id = ?").bind(Math.round(quantity), ingredientId).run();
  p.ingredient = { id: ing.id, name: ing.name, unit: ing.unit };
  return camelJson(p);
}
async function dailyPurchaseSpend(req: Request, env: Env) {
  const u = new URL(req.url), b: any[] = [];
  let s = "SELECT purchase_date AS date, SUM(total_cost) AS total_cost, COUNT(*) AS purchase_count FROM ingredient_purchases WHERE 1=1";
  if (u.searchParams.get('start')) { s += ' AND purchase_date >= ?'; b.push(u.searchParams.get('start')); }
  if (u.searchParams.get('end')) { s += ' AND purchase_date <= ?'; b.push(u.searchParams.get('end')); }
  s += ' GROUP BY purchase_date ORDER BY purchase_date DESC';
  return camelJson((await (b.length ? env.DB.prepare(s).bind(...b) : env.DB.prepare(s)).all()).results);
}

// ========== Dishes ==========
async function enrichDishes(list: any[], env: Env) {
  for (const d of list) {
    const ings = (await env.DB.prepare("SELECT di.*, i.name AS ing_name, i.unit AS ing_unit FROM dish_ingredients di LEFT JOIN ingredients i ON di.ingredient_id = i.id WHERE di.dish_id = ?").bind(d.id).all()).results;
    d.ingredients = ings.map((r: any) => ({ id: r.id, dishId: r.dish_id, ingredientId: r.ingredient_id, quantity: r.quantity, ingredient: r.ing_name ? { id: r.ingredient_id, name: r.ing_name, unit: r.ing_unit } : undefined }));
  }
  return list;
}
async function listDishes(req: Request, env: Env) {
  const o = new URL(req.url).searchParams.get('onShelf') === 'true';
  return camelJson(await enrichDishes(o ? (await env.DB.prepare('SELECT * FROM dishes WHERE on_shelf = 1 ORDER BY id DESC').all()).results : (await env.DB.prepare('SELECT * FROM dishes ORDER BY id DESC').all()).results, env));
}
async function getDishById(env: Env, id: number) {
  const d = await env.DB.prepare('SELECT * FROM dishes WHERE id = ?').bind(id).first() as any;
  if (!d) return fail(404, '菜品不存在');
  return camelJson((await enrichDishes([d], env))[0]);
}
async function createDish(req: Request, env: Env) {
  const b = await req.json() as any;
  if (!b.name?.trim()) return fail(400, '请填写名称');
  const d = await env.DB.prepare("INSERT INTO dishes (name, description, category, kind, cooking_method, image_url, on_shelf) VALUES (?,?,?,?,?,?,?) RETURNING *").bind(b.name.trim(), b.description||'', b.category||'', b.kind||'dish', b.cookingMethod||'', b.imageUrl||'', b.onShelf?1:0).first() as any;
  if (b.ingredients?.length) for (const di of b.ingredients) await env.DB.prepare("INSERT INTO dish_ingredients (dish_id, ingredient_id, quantity) VALUES (?,?,?)").bind(d.id, Number(di.ingredientId), Number(di.quantity)||0).run();
  return getDishById(env, d.id);
}
async function updateDish(req: Request, env: Env, id: number) {
  if (!await env.DB.prepare('SELECT 1 FROM dishes WHERE id = ?').bind(id).first()) return fail(404, '菜品不存在');
  const b = await req.json() as any;
  await env.DB.prepare("UPDATE dishes SET name=?, description=?, category=?, kind=?, cooking_method=?, image_url=?, on_shelf=?, updated_at=datetime('now') WHERE id=?").bind(b.name?.trim()??'', b.description??'', b.category??'', b.kind||'dish', b.cookingMethod??'', b.imageUrl??'', b.onShelf?1:0, id).run();
  await env.DB.prepare('DELETE FROM dish_ingredients WHERE dish_id = ?').bind(id).run();
  if (b.ingredients?.length) for (const di of b.ingredients) await env.DB.prepare("INSERT INTO dish_ingredients (dish_id, ingredient_id, quantity) VALUES (?,?,?)").bind(id, Number(di.ingredientId), Number(di.quantity)||0).run();
  return getDishById(env, id);
}
async function toggleShelf(env: Env, id: number) {
  const d = await env.DB.prepare('SELECT * FROM dishes WHERE id = ?').bind(id).first() as any;
  if (!d) return fail(404, '菜品不存在');
  await env.DB.prepare("UPDATE dishes SET on_shelf=?, updated_at=datetime('now') WHERE id=?").bind(d.on_shelf?0:1, id).run();
  return getDishById(env, id);
}
async function deleteDish(env: Env, id: number) {
  if (!await env.DB.prepare('SELECT 1 FROM dishes WHERE id = ?').bind(id).first()) return fail(404, '菜品不存在');
  await env.DB.prepare('DELETE FROM dish_ingredients WHERE dish_id = ?').bind(id).run();
  await env.DB.prepare('DELETE FROM dishes WHERE id = ?').bind(id).run();
  return ok();
}

// ========== Orders ==========
async function loadOrder(env: Env, id: number) {
  const o = await env.DB.prepare('SELECT * FROM orders WHERE id = ?').bind(id).first() as any;
  if (!o) return null;
  o.items = (await env.DB.prepare('SELECT * FROM order_items WHERE order_id = ? ORDER BY id').bind(id).all()).results;
  return o;
}
async function listOrders(env: Env) {
  const os = (await env.DB.prepare('SELECT * FROM orders ORDER BY id DESC').all()).results as any[];
  for (const o of os) o.items = (await env.DB.prepare('SELECT * FROM order_items WHERE order_id = ? ORDER BY id').bind(o.id).all()).results;
  return camelJson(os);
}
async function createOrder(req: Request, env: Env) {
  const b = await req.json() as { remark?: string; items: { dishId: number; quantity: number }[] };
  if (!b.items?.length) return fail(400, '订单不能为空');
  const needed: Record<number, number> = {};
  const ois: { dish_id: number; dish_name: string; quantity: number }[] = [];
  let valid = 0;
  for (const it of b.items) {
    if (it.quantity <= 0) continue;
    const dish = await env.DB.prepare('SELECT * FROM dishes WHERE id = ?').bind(it.dishId).first() as any;
    if (!dish) return fail(400, '菜品不存在');
    if (!dish.on_shelf) return fail(400, "菜品「" + dish.name + "」已下架");
    const ings = (await env.DB.prepare('SELECT * FROM dish_ingredients WHERE dish_id = ?').bind(it.dishId).all()).results as any[];
    for (const di of ings) needed[di.ingredient_id] = (needed[di.ingredient_id] || 0) + di.quantity * it.quantity;
    ois.push({ dish_id: dish.id, dish_name: dish.name, quantity: it.quantity });
    valid++;
  }
  if (!valid) return fail(400, '订单中所有菜品数量必须大于 0');
  for (const [idStr, qty] of Object.entries(needed)) {
    const ing = await env.DB.prepare('SELECT * FROM ingredients WHERE id = ?').bind(parseInt(idStr)).first() as any;
    if (!ing) return fail(400, '原材料不存在');
    if (ing.stock === STOCK_INFINITE) continue;
    if (ing.stock < qty) return fail(400, "原材料「" + ing.name + "」库存不足");
    await env.DB.prepare("UPDATE ingredients SET stock = stock - ?, updated_at = datetime('now') WHERE id = ?").bind(qty, parseInt(idStr)).run();
  }
  const order = await env.DB.prepare("INSERT INTO orders (status, remark) VALUES (?,?) RETURNING *").bind('pending', b.remark||'').first() as any;
  for (const oi of ois) await env.DB.prepare("INSERT INTO order_items (order_id, dish_id, dish_name, quantity) VALUES (?,?,?,?)").bind(order.id, oi.dish_id, oi.dish_name, oi.quantity).run();
  return camelJson(await loadOrder(env, order.id));
}
async function updateOrderStatus(req: Request, env: Env, id: number) {
  const b = await req.json() as { status: string };
  if (!VALID_STATUSES[b.status]) return fail(400, '无效的状态值');
  const order = await loadOrder(env, id);
  if (!order) return fail(404, '订单不存在');
  if (b.status === 'cancelled' && order.status !== 'cancelled') {
    const needed: Record<number, number> = {};
    for (const it of order.items) {
      const ings = (await env.DB.prepare('SELECT * FROM dish_ingredients WHERE dish_id = ?').bind(it.dish_id).all()).results as any[];
      for (const di of ings) needed[di.ingredient_id] = (needed[di.ingredient_id] || 0) + di.quantity * it.quantity;
    }
    for (const [idStr, qty] of Object.entries(needed)) {
      const ing = await env.DB.prepare('SELECT * FROM ingredients WHERE id = ?').bind(parseInt(idStr)).first() as any;
      if (!ing || ing.stock === STOCK_INFINITE) continue;
      await env.DB.prepare("UPDATE ingredients SET stock = stock + ?, updated_at = datetime('now') WHERE id = ?").bind(qty, parseInt(idStr)).run();
    }
  }
  await env.DB.prepare("UPDATE orders SET status = ?, updated_at = datetime('now') WHERE id = ?").bind(b.status, id).run();
  return camelJson(await loadOrder(env, id));
}
async function getOrderReceipt(env: Env, id: number) {
  const order = await loadOrder(env, id);
  if (!order) return fail(404, '订单不存在');
  const nm = shopName(env), items = order.items || [], total = items.reduce((s: number, it: any) => s + it.quantity, 0);
  const l: string[] = [centerText(nm, 32), centerText('订单小票', 32), '-'.repeat(32), '单号: #' + order.id, '时间: ' + order.created_at, '状态: ' + statusLabel(order.status), '-'.repeat(32)];
  for (const it of items) l.push(it.dish_name + '  x' + it.quantity);
  l.push('-'.repeat(32), '合计: ' + total + ' 件');
  if (order.remark?.trim()) l.push('备注: ' + order.remark);
  l.push('-'.repeat(32), centerText('谢谢惠顾', 32));
  const text = l.join('\n');
  return json({ receipt: { orderId: order.id, shopName: nm, status: order.status, remark: order.remark, items }, content: text });
}

// ========== Printer ==========
function printerStatusHandler() { return json({ ready: true, driver: 'stub', message: '模拟打印模式' }); }
function testPrintHandler() { return json({ success: true, jobId: 'stub-' + Date.now(), message: '已提交模拟打印', content: '测试打印 stub' }); }
async function printOrderHandler(env: Env, id: number) {
  const order = await loadOrder(env, id);
  if (!order) return fail(404, '订单不存在');
  const nm = shopName(env), items = order.items || [], total = items.reduce((s: number, it: any) => s + it.quantity, 0);
  const l: string[] = [centerText(nm, 32), centerText('订单小票', 32), '-'.repeat(32), '单号: #' + order.id, '时间: ' + order.created_at, '状态: ' + statusLabel(order.status), '-'.repeat(32)];
  for (const it of items) l.push(it.dish_name + '  x' + it.quantity);
  l.push('-'.repeat(32), '合计: ' + total + ' 件');
  if (order.remark?.trim()) l.push('备注: ' + order.remark);
  l.push('-'.repeat(32), centerText('谢谢惠顾', 32));
  return json({ success: true, jobId: 'stub-' + id + '-' + Date.now(), message: '已提交模拟打印', content: l.join('\n') });
}

// ========== Notify ==========
async function notifyStatusHandler(env: Env) {
  const c: string[] = [];
  if (env.FEISHU_WEBHOOK_URL) c.push('feishu');
  if (env.WECOM_WEBHOOK_URL) c.push('wecom');
  if (env.PUSHPLUS_TOKEN) c.push('pushplus');
  if (!c.length) return json({ enabled: false, channels: [], message: '未配置推送渠道' });
  return json({ enabled: true, channels: c, message: '新订单创建后将自动推送' });
}
async function sendNotify(env: Env, title: string, content: string) {
  const fn = async (u: string, p: any) => { try { const r = await fetch(u, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(p) }); if (!r.ok) { const t = await r.text(); console.error('notify fail', u, t.slice(0,200)); } } catch(e) { console.error('notify error', u, e); } };
  if (env.FEISHU_WEBHOOK_URL) fn(env.FEISHU_WEBHOOK_URL, { msg_type: 'text', content: { text: title + '\n\n' + content } });
  if (env.WECOM_WEBHOOK_URL) fn(env.WECOM_WEBHOOK_URL, { msgtype: 'text', text: { content: title + '\n\n' + content } });
  if (env.PUSHPLUS_TOKEN) fn('https://www.pushplus.plus/send', { token: env.PUSHPLUS_TOKEN, title, content: content.replace(/\n/g, '<br>') });
}
async function testNotifyHandler(env: Env) {
  const t = await notifyStatusHandler(env); const s: any = await t.json();
  if (!s.enabled) return fail(503, '未配置任何推送渠道');
  await sendNotify(env, '【' + shopName(env) + '】推送测试', '这是一条测试消息，说明订单推送渠道配置正常。');
  return json({ ok: true, message: '测试消息已发送' });
}
async function notifyOrderById(env: Env, id: number) {
  const order = await loadOrder(env, id);
  if (!order) return fail(404, '订单不存在');
  const items = order.items || [], total = items.reduce((s: number, it: any) => s + it.quantity, 0);
  let d = '单号: #' + order.id + '\n时间: ' + order.created_at + '\n状态: ' + statusLabel(order.status) + '\n—— 明细 ——\n';
  for (const it of items) d += '· ' + it.dish_name + ' × ' + it.quantity + '\n';
  d += '合计: ' + total + ' 件\n';
  if (order.remark?.trim()) d += '备注: ' + order.remark + '\n';
  await sendNotify(env, '【' + shopName(env) + '】新订单 #' + order.id, d);
  return json({ ok: true, message: '已推送' });
}

// ========== Upload ==========
async function uploadImage(req: Request, env: Env) {
  const fd = await req.formData(), file = fd.get('file') as File | null;
  if (!file) return fail(400, '未收到文件');
  const ext = (file.name?.split('.').pop()?.toLowerCase()) || 'png', key = Date.now() + '-' + Math.random().toString(36).slice(2, 8) + '.' + ext;
  await env.UPLOADS.put(key, await file.arrayBuffer(), { httpMetadata: { contentType: file.type || 'image/png' } });
  return json({ url: '/uploads/' + key });
}

// ========== AI ==========
async function aiCookingMethod(req: Request, env: Env) {
  const b = await req.json() as any;
  if (!b.name?.trim()) return fail(400, '请先填写名称');
  const kl = b.kind === 'soup' ? '汤品' : '菜品', it = b.ingredients?.length ? b.ingredients.join('、') : '（未提供原材料）';
  const r = await callAI(env, [{ role: 'system', content: '你是专业中餐厨师。' }, { role: 'user', content: '请为' + kl + '「' + b.name + '」给出烹饪方法。\n原材料：' + it + '。\n要求：分步骤，6步以内，只输出步骤。' }]);
  return json({ cookingMethod: r });
}
async function aiCookingMethodFromURL(req: Request, env: Env) {
  const b = await req.json() as any;
  let url = (b.url || '').trim();
  if (!url) return fail(400, '请填写来源网址');
  if (!url.startsWith('http')) url = 'https://' + url;
  const res = await fetch(url, { headers: { 'User-Agent': 'Mozilla/5.0' } });
  if (!res.ok) return fail(502, '抓取网址失败: HTTP ' + res.status);
  let text = await res.text();
  text = text.replace(/<script[^>]*>[\s\S]*?<\/script>/gi, ' ').replace(/<style[^>]*>[\s\S]*?<\/style>/gi, ' ').replace(/<[^>]+>/g, ' ').replace(/&[^;]+;/g, ' ').replace(/\s+/g, ' ').trim();
  if (text.length > 8000) text = text.slice(0, 8000);
  if (!text) return fail(502, '未能提取到内容');
  const nh = b.name?.trim() ? '目标菜品：「' + b.name + '」。\n' : '';
  const r = await callAI(env, [{ role: 'system', content: '你擅长从网页文本提炼食谱步骤。' }, { role: 'user', content: nh + '从以下网页内容提取烹饪方法，分步骤，8步以内，只输出步骤。\n\n' + text }]);
  return json({ cookingMethod: r });
}
async function callAI(env: Env, msgs: { role: string; content: string }[]) {
  const k = env.AI_API_KEY; if (!k) throw new Error('未配置 AI_API_KEY');
  let bu = (env.AI_BASE_URL || 'https://api.openai.com/v1').replace(/\/+$/, '');
  const m = env.AI_MODEL || 'gpt-4o-mini';
  let ep = bu; if (!ep.includes('/chat/completions')) ep += ep.endsWith('/v1') ? '/chat/completions' : '/v1/chat/completions';
  const r = await fetch(ep, { method: 'POST', headers: { 'Content-Type': 'application/json', 'Authorization': 'Bearer ' + k }, body: JSON.stringify({ model: m, temperature: 0.7, messages: msgs }) });
  if (!r.ok) { const t = await r.text(); throw new Error('AI 错误: ' + r.status + ': ' + t.slice(0,300)); }
  const j: any = await r.json();
  if (j.error) throw new Error(j.error.message);
  return (j.choices?.[0]?.message?.content || '').trim();
}

// ========== Router ==========
async function handleApi(request: Request, env: Env) {
  const url = new URL(request.url), path = url.pathname.replace(/\/$/, ''), method = request.method;
  const r = parseRoute(path);
  if (!r) return fail(404, 'not found');
  const { resource, id, suffix } = r;

  try {
    if (resource === 'ingredients') {
      if (!id) { return method === 'GET' ? listIngredients(env) : method === 'POST' ? createIngredient(request, env) : fail(405, ''); }
      if (id && suffix === 'stock' && method === 'PATCH') return setIngredientStock(request, env, id);
      if (id && method === 'PUT') return updateIngredient(request, env, id);
      if (id && method === 'DELETE') return deleteIngredient(env, id);
    } else if (resource === 'finance') {
      if (suffix === 'purchases' && !id) { return method === 'GET' ? listPurchases(request, env) : method === 'POST' ? createPurchase(request, env) : fail(405, ''); }
      if (suffix === 'daily-spend' && method === 'GET') return dailyPurchaseSpend(request, env);
    } else if (resource === 'dishes') {
      if (!id) { return method === 'GET' ? listDishes(request, env) : method === 'POST' ? createDish(request, env) : fail(405, ''); }
      if (id && suffix === 'shelf' && method === 'PATCH') return toggleShelf(env, id);
      if (id && method === 'GET') return getDishById(env, id);
      if (id && method === 'PUT') return updateDish(request, env, id);
      if (id && method === 'DELETE') return deleteDish(env, id);
    } else if (resource === 'orders') {
      if (!id) { return method === 'GET' ? listOrders(env) : method === 'POST' ? createOrder(request, env) : fail(405, ''); }
      if (id && suffix === 'status' && method === 'PATCH') return updateOrderStatus(request, env, id);
      if (id && suffix === 'receipt' && method === 'GET') return getOrderReceipt(env, id);
      if (id && suffix === 'print' && method === 'POST') return printOrderHandler(env, id);
      if (id && suffix === 'notify' && method === 'POST') return notifyOrderById(env, id);
    } else if (resource === 'printer') {
      if (suffix === 'status') return printerStatusHandler();
      if (suffix === 'test' && method === 'POST') return testPrintHandler();
    } else if (resource === 'notify') {
      if (suffix === 'status') return notifyStatusHandler(env);
      if (suffix === 'test' && method === 'POST') return testNotifyHandler(env);
    } else if (resource === 'upload' && method === 'POST') {
      return uploadImage(request, env);
    } else if (resource === 'ai') {
      if (suffix === 'cooking-method' && method === 'POST') return aiCookingMethod(request, env);
      if (suffix === 'cooking-method-from-url' && method === 'POST') return aiCookingMethodFromURL(request, env);
    }
  } catch (e: any) {
    return fail(500, e.message || 'Internal Server Error');
  }
  return fail(404, 'not found');
}

export async function onRequest(context: any): Promise<Response> {
  const { request, env } = context;
  if (request.method === 'OPTIONS') return new Response(null, { status: 204, headers: corsHeaders() });
  const url = new URL(request.url);
  if (url.pathname.startsWith('/api/')) return handleApi(request, env);
  return fail(404, 'not found');
}
