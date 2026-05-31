// REST 接口定义（供可视化文档页使用）
export const apiGroups = [
  {
    name: '菜品 / 汤品',
    endpoints: [
      {
        method: 'GET',
        path: '/api/dishes',
        desc: '获取菜品列表',
        query: [{ name: 'onShelf', example: 'true', note: '可选，仅返回已上架' }],
        response: '[{ "id": 1, "name": "西红柿炒鸡蛋", "kind": "dish", "cookingMethod": "...", "onShelf": true, "ingredients": [...] }]',
      },
      {
        method: 'GET',
        path: '/api/dishes/:id',
        desc: '获取单个菜品详情',
        response: '{ "id": 1, "name": "...", "ingredients": [...] }',
      },
      {
        method: 'POST',
        path: '/api/dishes',
        desc: '新增菜品或汤品',
        body: '{ "name": "紫菜蛋花汤", "kind": "soup", "category": "汤品", "description": "...", "cookingMethod": "...", "imageUrl": "/uploads/xxx.png", "onShelf": true, "ingredients": [{ "ingredientId": 1, "quantity": 2 }] }',
      },
      {
        method: 'PUT',
        path: '/api/dishes/:id',
        desc: '更新菜品（含重建原材料关联）',
        body: '同 POST',
      },
      {
        method: 'PATCH',
        path: '/api/dishes/:id/shelf',
        desc: '切换上架 / 下架',
        response: '{ "id": 1, "onShelf": false, ... }',
      },
      {
        method: 'DELETE',
        path: '/api/dishes/:id',
        desc: '删除菜品',
        response: '{ "ok": true }',
      },
    ],
  },
  {
    name: '原材料 / 库存',
    endpoints: [
      {
        method: 'GET',
        path: '/api/ingredients',
        desc: '原材料列表',
        response: '[{ "id": 1, "name": "鸡蛋", "unit": "个", "stock": 100 }]',
      },
      {
        method: 'POST',
        path: '/api/ingredients',
        desc: '新增原材料',
        body: '{ "name": "豆瓣酱", "unit": "克", "stock": -1 }',
        note: 'stock = -1 表示无限库存',
      },
      {
        method: 'PUT',
        path: '/api/ingredients/:id',
        desc: '更新原材料',
        body: '{ "name": "...", "unit": "...", "stock": 50 }',
      },
      {
        method: 'PATCH',
        path: '/api/ingredients/:id/stock',
        desc: '设置库存',
        body: '{ "stock": 100, "infinite": false } 或 { "infinite": true }',
      },
      {
        method: 'DELETE',
        path: '/api/ingredients/:id',
        desc: '删除原材料',
        response: '{ "ok": true }',
      },
    ],
  },
  {
    name: '订单',
    endpoints: [
      {
        method: 'GET',
        path: '/api/orders',
        desc: '订单列表（含明细）',
        response: '[{ "id": 1, "status": "pending", "items": [...], "remark": "" }]',
      },
      {
        method: 'POST',
        path: '/api/orders',
        desc: '创建订单（扣减库存）',
        body: '{ "remark": "不要香菜", "items": [{ "dishId": 1, "quantity": 2 }] }',
      },
      {
        method: 'PATCH',
        path: '/api/orders/:id/status',
        desc: '更新订单状态',
        body: '{ "status": "paid" }',
        note: 'status: pending | paid | done | cancelled',
      },
    ],
  },
  {
    name: '上传 & AI',
    endpoints: [
      {
        method: 'POST',
        path: '/api/upload',
        desc: '上传图片（multipart，字段名 file）',
        body: 'multipart/form-data: file=<图片文件>',
        response: '{ "url": "/uploads/1234567890.png" }',
      },
      {
        method: 'POST',
        path: '/api/ai/cooking-method',
        desc: 'AI 生成烹饪方法（需配置 AI_API_KEY）',
        body: '{ "name": "麻婆豆腐", "kind": "dish", "ingredients": ["豆腐", "牛肉末"] }',
        response: '{ "cookingMethod": "1. ...\\n2. ..." }',
      },
      {
        method: 'POST',
        path: '/api/ai/cooking-method-from-url',
        desc: '从来源网站抓取并提炼烹饪方法',
        body: '{ "name": "番茄炒蛋", "url": "https://example.com/recipe" }',
        response: '{ "cookingMethod": "1. ...\\n2. ..." }',
      },
    ],
  },
]

export const methodColor = {
  GET: '#2ecc71',
  POST: '#3498db',
  PUT: '#f39c12',
  PATCH: '#9b59b6',
  DELETE: '#e74c3c',
}
