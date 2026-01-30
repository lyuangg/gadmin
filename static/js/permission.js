/**
 * 权限管理模块
 * 负责权限数据的获取、缓存和检查
 */

(function() {
    'use strict';

    // 权限管理器
    window.PermissionManager = {
        // 权限缓存
        permissions: [],
        isSuperAdmin: false,
        initialized: false,
        initPromise: null,
        
        // localStorage 缓存 key
        CACHE_KEY: 'user_permissions_cache',

        // 菜单权限映射配置
        menuPermissions: {
            '/admin': null,  // 首页不需要权限
            '/admin/users': {
                path: '/admin/api/users',
                method: 'GET'
            },
            '/admin/roles': {
                path: '/admin/api/roles',
                method: 'GET'
            },
            '/admin/permissions': {
                path: '/admin/api/permissions',
                method: 'GET'
            },
            '/admin/dictionaries': {
                path: '/admin/api/dictionaries/types',
                method: 'GET'
            }
        },

        // 按钮权限映射配置（按页面分组）
        buttonPermissions: {
            '/admin/users': {
                'add': { path: '/admin/api/users', method: 'POST' },
                'edit': { path: '/admin/api/users/:id', method: 'PUT' },
                'delete': { path: '/admin/api/users/:id', method: 'DELETE' },
                'toggleStatus': { path: '/admin/api/users/:id/toggle-status', method: 'PUT' },
                'resetPassword': { path: '/admin/api/users/:id/reset-password', method: 'POST' }
            },
            '/admin/roles': {
                'add': { path: '/admin/api/roles', method: 'POST' },
                'edit': { path: '/admin/api/roles/:id', method: 'PUT' },
                'delete': { path: '/admin/api/roles/:id', method: 'DELETE' },
                'assignPermissions': { path: '/admin/api/roles/:id/permissions', method: 'PUT' }
            },
            '/admin/permissions': {
                'add': { path: '/admin/api/permissions', method: 'POST' },
                'edit': { path: '/admin/api/permissions/:id', method: 'PUT' },
                'delete': { path: '/admin/api/permissions/:id', method: 'DELETE' }
            },
            '/admin/dictionaries': {
                'add': { path: '/admin/api/dictionaries/types', method: 'POST' },
                'edit': { path: '/admin/api/dictionaries/types/:id', method: 'PUT' },
                'delete': { path: '/admin/api/dictionaries/types/:id', method: 'DELETE' },
                'addItem': { path: '/admin/api/dictionaries/items', method: 'POST' },
                'editItem': { path: '/admin/api/dictionaries/items/:id', method: 'PUT' },
                'deleteItem': { path: '/admin/api/dictionaries/items/:id', method: 'DELETE' }
            }
        },

        /**
         * 从 JWT token 中解析 claims（不验证签名，仅解析 payload）
         * @param {string} token - JWT token
         * @returns {object|null} 解析后的 claims 对象，失败返回 null
         */
        parseTokenClaims: function(token) {
            if (!token) {
                return null;
            }
            
            try {
                // JWT 格式: header.payload.signature
                var parts = token.split('.');
                if (parts.length !== 3) {
                    return null;
                }
                
                // 解析 payload（base64url 解码）
                var payload = parts[1];
                // 添加 padding（base64url 可能缺少 =）
                while (payload.length % 4) {
                    payload += '=';
                }
                // 将 base64url 转换为 base64
                payload = payload.replace(/-/g, '+').replace(/_/g, '/');
                
                // 解码 base64
                var decodedPayload = atob(payload);
                var claims = JSON.parse(decodedPayload);
                
                return claims;
            } catch (error) {
                console.error('解析 token 失败:', error);
                return null;
            }
        },

        /**
         * 从 localStorage 读取缓存的权限数据
         * @returns {object|null} 缓存的权限数据，失败返回 null
         */
        loadFromCache: function() {
            try {
                var cached = localStorage.getItem(this.CACHE_KEY);
                if (!cached) {
                    return null;
                }
                
                var cacheData = JSON.parse(cached);
                // 检查缓存是否有效（需要包含必要字段）
                if (cacheData && typeof cacheData === 'object' && 
                    typeof cacheData.is_super_admin !== 'undefined' && 
                    Array.isArray(cacheData.permissions)) {
                    
                    // 验证缓存是否与当前 token 匹配（通过解析 token 获取 user_id）
                    var token = getToken();
                    if (token) {
                        var claims = this.parseTokenClaims(token);
                        if (claims && cacheData.user_id && cacheData.user_id !== claims.user_id) {
                            // token 中的 user_id 与缓存不匹配，说明切换了账号，清除缓存
                            console.log('检测到账号切换，清除旧缓存');
                            this.clearCache();
                            return null;
                        }
                    }
                    
                    return cacheData;
                }
                return null;
            } catch (error) {
                console.error('读取权限缓存失败:', error);
                return null;
            }
        },

        /**
         * 将权限数据保存到 localStorage
         * @param {object} data - 权限数据 {is_super_admin: boolean, permissions: array}
         */
        saveToCache: function(data) {
            try {
                // 从 token 中获取 user_id，用于验证缓存是否属于当前用户
                var userID = null;
                var token = getToken();
                if (token) {
                    var claims = this.parseTokenClaims(token);
                    if (claims && claims.user_id) {
                        userID = claims.user_id;
                    }
                }
                
                var cacheData = {
                    is_super_admin: data.is_super_admin === true,
                    permissions: data.permissions || [],
                    user_id: userID, // 保存 user_id 用于验证
                    cached_at: Date.now() // 缓存时间戳
                };
                localStorage.setItem(this.CACHE_KEY, JSON.stringify(cacheData));
            } catch (error) {
                console.error('保存权限缓存失败:', error);
            }
        },

        /**
         * 清除权限缓存
         */
        clearCache: function() {
            try {
                localStorage.removeItem(this.CACHE_KEY);
            } catch (error) {
                console.error('清除权限缓存失败:', error);
            }
        },

        /**
         * 初始化权限（页面加载时调用）
         * 只从缓存或 token 中读取，不主动请求接口
         * @returns {Promise} 返回 Promise，权限加载完成后 resolve
         */
        initPermissions: function() {
            // 如果已经初始化，直接返回缓存的 Promise
            if (this.initialized && this.initPromise) {
                return this.initPromise;
            }

            // 如果正在初始化，返回同一个 Promise
            if (this.initPromise) {
                return this.initPromise;
            }

            var self = this;
            
            // 首先尝试从 localStorage 缓存读取
            var cachedData = this.loadFromCache();
            if (cachedData) {
                this.isSuperAdmin = cachedData.is_super_admin === true;
                this.permissions = cachedData.permissions || [];
                this.initialized = true;
                console.log('从缓存加载权限:', {
                    isSuperAdmin: this.isSuperAdmin,
                    permissionsCount: this.permissions.length
                });
                this.initPromise = Promise.resolve(true);
                return this.initPromise;
            }
            
            // 缓存不存在，尝试从 token 中解析超级管理员状态
            var token = getToken();
            if (token) {
                var claims = this.parseTokenClaims(token);
                if (claims && claims.is_super_admin === true) {
                    // 超级管理员，直接设置状态，不需要请求接口
                    this.isSuperAdmin = true;
                    this.permissions = [];
                    this.initialized = true;
                    // 缓存超级管理员状态
                    this.saveToCache({ is_super_admin: true, permissions: [] });
                    console.log('从 token 中检测到超级管理员');
                    this.initPromise = Promise.resolve(true);
                    return this.initPromise;
                }
            }

            // 既没有缓存，也不是超级管理员，设置为未初始化状态
            // 等待登录时主动调用 fetchAndCachePermissions 获取权限
            this.isSuperAdmin = false;
            this.permissions = [];
            this.initialized = false;
            console.log('未找到权限缓存，等待登录时获取');
            this.initPromise = Promise.resolve(false);
            return this.initPromise;
        },

        /**
         * 匹配权限路径（支持路径参数，如 :id）
         * @param {string} permPath - 权限路径（如 /admin/api/users/:id）
         * @param {string} reqPath - 请求路径（如 /admin/api/users/1）
         * @returns {boolean}
         */
        matchPermissionPath: function(permPath, reqPath) {
            // 1. 精确匹配
            if (permPath === reqPath) {
                return true;
            }

            // 2. 路径参数匹配（将 :id 替换为正则表达式）
            // 转义特殊字符，然后将 :参数名 替换为 [^/]+
            var pattern = permPath.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'); // 转义特殊字符
            pattern = pattern.replace(/:\w+/g, '[^/]+'); // 将 :id 替换为 [^/]+
            var regex = new RegExp('^' + pattern + '$');
            if (regex.test(reqPath)) {
                return true;
            }

            // 3. 路径前缀匹配（支持 /* 后缀）
            if (permPath.endsWith('/*')) {
                var prefix = permPath.slice(0, -2);
                if (reqPath.startsWith(prefix + '/')) {
                    return true;
                }
            }

            return false;
        },

        /**
         * 检查是否有某个权限
         * @param {string} path - 权限路径
         * @param {string} method - 请求方法（GET/POST/PUT/DELETE等）
         * @returns {boolean}
         */
        hasPermission: function(path, method) {
            // 超级管理员拥有所有权限
            if (this.isSuperAdmin) {
                return true;
            }

            // 遍历权限列表，查找匹配的权限
            for (var i = 0; i < this.permissions.length; i++) {
                var perm = this.permissions[i];
                // 方法必须匹配（不区分大小写）
                // 注意：如果权限的 method 为空，则匹配所有方法
                if (perm.method && perm.method.toUpperCase() !== method.toUpperCase()) {
                    continue;
                }
                // 路径匹配
                if (this.matchPermissionPath(perm.path, path)) {
                    return true;
                }
            }

            return false;
        },

        /**
         * 检查菜单是否可见
         * @param {string} menuPath - 菜单路径（如 /admin/users）
         * @returns {boolean}
         */
        isMenuVisible: function(menuPath) {
            var menuPerm = this.menuPermissions[menuPath];
            
            // 如果菜单不需要权限，始终显示
            if (!menuPerm) {
                return true;
            }

            // 检查是否有对应的权限
            var hasPerm = this.hasPermission(menuPerm.path, menuPerm.method);
            console.log('检查菜单权限:', {
                menuPath: menuPath,
                requiredPath: menuPerm.path,
                requiredMethod: menuPerm.method,
                hasPermission: hasPerm,
                isSuperAdmin: this.isSuperAdmin
            });
            return hasPerm;
        },

        /**
         * 检查按钮是否可见
         * @param {string} pagePath - 页面路径（如 /admin/users）
         * @param {string} buttonKey - 按钮标识（如 'add', 'edit', 'delete'）
         * @returns {boolean}
         */
        isButtonVisible: function(pagePath, buttonKey) {
            // 超级管理员拥有所有权限
            if (this.isSuperAdmin) {
                return true;
            }

            var pageButtons = this.buttonPermissions[pagePath];
            if (!pageButtons) {
                return false;
            }

            var buttonPerm = pageButtons[buttonKey];
            if (!buttonPerm) {
                return false;
            }

            // 检查是否有对应的权限
            return this.hasPermission(buttonPerm.path, buttonPerm.method);
        },

        /**
         * 刷新权限（重新从服务器获取并更新缓存）
         * @returns {Promise}
         */
        refreshPermissions: function() {
            // 清除缓存
            this.clearCache();
            this.initialized = false;
            this.initPromise = null;
            return this.initPermissions();
        },

        /**
         * 在登录成功后调用，主动获取并缓存权限
         * @returns {Promise}
         */
        fetchAndCachePermissions: function() {
            var self = this;
            // 清除旧缓存
            this.clearCache();
            this.initialized = false;
            this.initPromise = null;
            
            // 从 token 中解析超级管理员状态
            var token = getToken();
            if (token) {
                var claims = this.parseTokenClaims(token);
                if (claims && claims.is_super_admin === true) {
                    // 超级管理员，直接设置状态并缓存
                    this.isSuperAdmin = true;
                    this.permissions = [];
                    this.initialized = true;
                    this.saveToCache({ is_super_admin: true, permissions: [] });
                    console.log('登录成功：从 token 检测到超级管理员，已缓存');
                    this.initPromise = Promise.resolve(true);
                    return this.initPromise;
                }
            }
            
            // 非超级管理员，请求接口获取权限并缓存
            this.initPromise = api.auth.getUserPermissions()
                .then(function(response) {
                    var data = response && response.data !== undefined ? response.data : response;
                    
                    if (data && typeof data === 'object') {
                        self.isSuperAdmin = data.is_super_admin === true;
                        self.permissions = data.permissions || [];
                        self.initialized = true;
                        // 保存到缓存
                        self.saveToCache({
                            is_super_admin: self.isSuperAdmin,
                            permissions: self.permissions
                        });
                        console.log('登录成功：权限已获取并缓存', {
                            isSuperAdmin: self.isSuperAdmin,
                            permissionsCount: self.permissions.length
                        });
                        return true;
                    } else {
                        console.error('获取权限失败: 数据格式错误', data);
                        return false;
                    }
                })
                .catch(function(error) {
                    console.error('获取权限失败:', error);
                    return false;
                });
            
            return this.initPromise;
        }
    };

    // 页面加载时只从缓存读取权限，不主动请求接口
    // 权限接口请求只在登录成功后由登录页面主动调用 fetchAndCachePermissions
    // 这里只做初始化，从缓存或 token 中读取
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', function() {
            window.PermissionManager.initPermissions();
        });
    } else {
        // DOM 已经加载完成，立即初始化（只从缓存读取，不请求接口）
        window.PermissionManager.initPermissions();
    }
})();
