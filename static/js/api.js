/**
 * API 请求封装
 * 依赖：main.js 中的 getToken() 和 getCookie() 函数
 */

// 错误提示去重：记录已显示的错误提示，避免重复显示
var errorMessageCache = {
    forbidden: {
        shown: false,
        timestamp: 0
    }
};

// 清除错误提示缓存的辅助函数（可选，用于重置状态）
function clearErrorMessageCache() {
    errorMessageCache.forbidden.shown = false;
    errorMessageCache.forbidden.timestamp = 0;
}

// 创建 axios 实例
var api = {
    // 基础请求方法
    request: function(config) {
        // 如果是需要认证的请求（以 /admin/api/ 开头），自动添加 token
        if (config.url && config.url.startsWith('/admin/api/')) {
            var token = getToken();
            if (token) {
                if (!config.headers) {
                    config.headers = {};
                }
                config.headers['Authorization'] = 'Bearer ' + token;
            }
        }
        
        return axios(config).then(function(response) {
            // 统一处理响应格式 {code, msg, data}
            var data = response.data;
            if (data && typeof data.code !== 'undefined') {
                // 如果 code 不为 0，表示有错误
                if (data.code !== 0) {
                    var error = new Error(data.msg || '请求失败');
                    error.response = {
                        data: {
                            code: data.code,
                            msg: data.msg,
                            data: data.data
                        }
                    };
                    return Promise.reject(error);
                }
                // 成功时返回 data 字段
                response.data = data.data;
            }
            return response;
        }).catch(function(error) {
            // 统一处理业务错误码（统一响应格式 {code, msg, data}）
            var errorCode = null;
            var errorMsg = '';
            
            // 检查响应数据中的业务错误码
            if (error.response) {
                var data = error.response.data;
                // 统一响应格式 {code, msg, data}（在 then 中已经处理过，这里 error.response.data 就是 {code, msg, data}）
                if (data && typeof data.code !== 'undefined') {
                    errorCode = data.code;
                    errorMsg = data.msg || '';
                }
                // 兼容直接返回 {error: "message"} 的格式（HTTP 状态码错误，如中间件直接返回的）
                else if (data && data.error) {
                    errorMsg = data.error;
                    // HTTP 401 或 403 状态码也视为对应的业务错误码
                    if (error.response.status === 401) {
                        errorCode = 401;
                    } else if (error.response.status === 403) {
                        errorCode = 403;
                    }
                }
            }
            
            // 根据业务错误码处理
            if (errorCode === 401) {
                // 401: 未认证（Token 无效、过期、失效等）
                var msg401 = errorMsg || '登录已过期';
                msg401 += '，确定重新登录';
                
                // 使用确认框，用户点击确定后才跳转
                var handle401 = function() {
                    // 清除 token
                    localStorage.removeItem('token');
                    localStorage.removeItem('user');
                    document.cookie = 'token=; path=/; max-age=0';
                    // 跳转到登录页
                    window.location.href = '/login';
                };
                
                // 使用 Element Plus 的 MessageBox
                ElMessageBox.confirm(msg401, '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning',
                }).then(function() {
                    // 用户点击确定
                    handle401();
                }).catch(function() {
                    // 用户点击取消，什么也不做
                });
                // 返回一个永远不会 resolve 或 reject 的 Promise，完全阻止后续的错误处理
                // 这样调用方的 .catch() 不会执行
                return new Promise(function() {
                    // 这个 Promise 永远不会 resolve 或 reject，保持挂起状态
                });
            } else if (errorCode === 403) {
                // 403: 无权限（权限不足、用户被禁用等）
                // 只弹出提示，不跳转
                // 使用去重机制，3秒内只显示一次提示
                var now = Date.now();
                var cache = errorMessageCache.forbidden;
                var timeWindow = 3000; // 3秒时间窗口
                
                // 如果已经显示过，且在时间窗口内，则不重复显示
                if (cache.shown && (now - cache.timestamp) < timeWindow) {
                    return new Promise(function() {
                        // 这个 Promise 永远不会 resolve 或 reject，保持挂起状态
                    });
                }
                
                // 标记为已显示，记录时间戳
                cache.shown = true;
                cache.timestamp = now;
                
                var msg403 = errorMsg || '没有权限访问此资源';
                if (typeof ElMessage !== 'undefined') {
                    ElMessage.error(msg403);
                } else if (typeof window.ElMessage !== 'undefined') {
                    window.ElMessage.error(msg403);
                } else {
                    alert(msg403);
                }
                
                // 3秒后重置标志，允许再次显示（如果还有新的 403 错误）
                setTimeout(function() {
                    cache.shown = false;
                }, timeWindow);
                
                return new Promise(function() {
                    // 这个 Promise 永远不会 resolve 或 reject，保持挂起状态
                });
            }
            
            // 其他错误直接抛出
            return Promise.reject(error);
        });
    },
    
    // GET 请求
    get: function(url, config) {
        var requestConfig = Object.assign({}, config || {}, {
            method: 'GET',
            url: url
        });
        return this.request(requestConfig);
    },
    
    // POST 请求
    post: function(url, data, config) {
        var requestConfig = Object.assign({}, config || {}, {
            method: 'POST',
            url: url,
            data: data
        });
        return this.request(requestConfig);
    },
    
    // PUT 请求
    put: function(url, data, config) {
        var requestConfig = Object.assign({}, config || {}, {
            method: 'PUT',
            url: url,
            data: data
        });
        return this.request(requestConfig);
    },
    
    // DELETE 请求
    delete: function(url, config) {
        var requestConfig = Object.assign({}, config || {}, {
            method: 'DELETE',
            url: url
        });
        return this.request(requestConfig);
    },
    
    // PATCH 请求
    patch: function(url, data, config) {
        var requestConfig = Object.assign({}, config || {}, {
            method: 'PATCH',
            url: url,
            data: data
        });
        return this.request(requestConfig);
    },
    
    // ==================== 业务 API 方法 ====================
    
    /**
     * 权限管理 API
     */
    permissions: {
        // 获取权限列表（支持分页和筛选）
        getList: function(params) {
            return api.get('/admin/api/permissions', { params: params });
        },
        // 创建权限
        create: function(data) {
            return api.post('/admin/api/permissions', data);
        },
        // 更新权限
        update: function(id, data) {
            return api.put('/admin/api/permissions/' + id, data);
        },
        // 删除权限
        delete: function(id) {
            return api.delete('/admin/api/permissions/' + id);
        },
        // 批量删除权限
        batchDelete: function(ids) {
            return api.post('/admin/api/permissions/batch-delete', {
                ids: ids
            });
        }
    },
    
    /**
     * 角色管理 API
     */
    roles: {
        // 获取角色列表（支持分页和排序）
        getList: function(params) {
            return api.get('/admin/api/roles', { params: params || {} });
        },
        // 创建角色
        create: function(data) {
            return api.post('/admin/api/roles', data);
        },
        // 更新角色
        update: function(id, data) {
            return api.put('/admin/api/roles/' + id, data);
        },
        // 删除角色
        delete: function(id) {
            return api.delete('/admin/api/roles/' + id);
        },
        // 分配权限
        assignPermissions: function(id, permissionIds) {
            return api.put('/admin/api/roles/' + id + '/permissions', {
                permission_ids: permissionIds
            });
        }
    },
    
    /**
     * 用户管理 API
     */
    users: {
        // 获取用户列表（支持分页和筛选）
        getList: function(params) {
            return api.get('/admin/api/users', { params: params });
        },
        // 创建用户
        create: function(data) {
            return api.post('/admin/api/users', data);
        },
        // 更新用户
        update: function(id, data) {
            return api.put('/admin/api/users/' + id, data);
        },
        // 删除用户
        delete: function(id) {
            return api.delete('/admin/api/users/' + id);
        },
        // 重置密码
        resetPassword: function(id) {
            return api.post('/admin/api/users/' + id + '/reset-password');
        },
        // 切换用户状态（启用/禁用）
        toggleStatus: function(id) {
            return api.put('/admin/api/users/' + id + '/toggle-status', {});
        }
    },

    /**
     * 字典管理 API
     */
    dictionaries: {
        // 字典类型
        getTypes: function(params) {
            return api.get('/admin/api/dictionaries/types', { params: params || {} });
        },
        createType: function(data) {
            return api.post('/admin/api/dictionaries/types', data);
        },
        updateType: function(id, data) {
            return api.put('/admin/api/dictionaries/types/' + id, data);
        },
        deleteType: function(id) {
            return api.delete('/admin/api/dictionaries/types/' + id);
        },
        // 字典项
        getItems: function(params) {
            return api.get('/admin/api/dictionaries/items', { params: params || {} });
        },
        getItemsByCode: function(code) {
            return api.get('/admin/api/dictionaries/items/by-code', { params: { code: code } });
        },
        createItem: function(data) {
            return api.post('/admin/api/dictionaries/items', data);
        },
        updateItem: function(id, data) {
            return api.put('/admin/api/dictionaries/items/' + id, data);
        },
        deleteItem: function(id) {
            return api.delete('/admin/api/dictionaries/items/' + id);
        }
    },

    /**
     * 操作日志 API
     */
    operationLogs: {
        // 获取操作日志列表（支持分页和筛选）
        getList: function(params) {
            return api.get('/admin/api/operation-logs', { params: params });
        }
    },
    
    /**
     * 个人资料 API
     */
    profile: {
        // 修改密码
        changePassword: function(oldPassword, newPassword) {
            return api.put('/admin/api/profile/password', {
                old_password: oldPassword,
                new_password: newPassword
            });
        },
        // 更新头像
        updateAvatar: function(avatarURL) {
            return api.put('/admin/api/profile/avatar', {
                avatar: avatarURL
            });
        }
    },
    
    /**
     * 认证 API
     */
    auth: {
        // 退出登录
        logout: function() {
            return api.post('/admin/api/logout', {});
        },
        // 获取当前用户权限
        getUserPermissions: function() {
            return api.get('/admin/api/user/permissions');
        }
    }
};
