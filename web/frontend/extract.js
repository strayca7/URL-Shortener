// 使用 fetch API 进行 HTTP 请求

// 登录函数，发送登录请求并处理响应
async function login(username, password) {
  try {
    const response = await fetch('/api/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      throw new Error('登录失败');
    }

    const data = await response.json();

    // 从响应中提取令牌
    const accessToken = data.access_token;
    const refreshToken = data.refresh_token;

    // 将令牌存储在本地存储中
    localStorage.setItem('accessToken', accessToken);
    localStorage.setItem('refreshToken', refreshToken);

    console.log('登录成功');
  } catch (error) {
    console.error(error);
  }
}

// 获取受保护资源的函数，使用存储的令牌
async function fetchProtectedResource() {
  try {
    const accessToken = localStorage.getItem('accessToken');

    if (!accessToken) {
      throw new Error('未找到访问令牌，请先登录');
    }

    const response = await fetch('/api/protected-resource', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    });

    if (response.status === 401) {
      // 访问令牌可能已过期，尝试使用刷新令牌获取新的访问令牌
      await refreshAccessToken();
      // 重新调用获取受保护资源的函数
      return fetchProtectedResource();
    }

    if (!response.ok) {
      throw new Error('获取受保护资源失败');
    }

    const data = await response.json();
    console.log('受保护资源数据:', data);
  } catch (error) {
    console.error(error);
  }
}

// 使用刷新令牌获取新的访问令牌
async function refreshAccessToken() {
  try {
    const refreshToken = localStorage.getItem('refreshToken');

    if (!refreshToken) {
      throw new Error('未找到刷新令牌，请重新登录');
    }

    const response = await fetch('/api/refresh-token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      throw new Error('刷新访问令牌失败');
    }

    const data = await response.json();
    const newAccessToken = data.access_token;

    // 更新本地存储中的访问令牌
    localStorage.setItem('accessToken', newAccessToken);
  } catch (error) {
    console.error(error);
    // 刷新令牌失败，可能需要重新登录
  }
}

// 示例使用
(async () => {
  await login('your_username', 'your_password');
  await fetchProtectedResource();
})();
