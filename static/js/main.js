/**
 * 通用工具函数
 */

// 获取 token 的辅助函数
function getToken() {
    // 先从 localStorage 获取
    const localToken = localStorage.getItem('token');
    if (localToken) {
        return localToken;
    }
    
    // 如果 localStorage 没有，从 cookie 获取
    return getCookie('token');
}

// 获取 cookie 的辅助函数
function getCookie(name) {
    const value = '; ' + document.cookie;
    const parts = value.split('; ' + name + '=');
    if (parts.length === 2) {
        return parts.pop().split(';').shift();
    }
    return null;
}

// 设置 cookie 的辅助函数
function setCookie(name, value, days) {
    const expires = days ? '; expires=' + new Date(Date.now() + days * 24 * 60 * 60 * 1000).toUTCString() : '';
    document.cookie = name + '=' + value + expires + '; path=/';
}

// 退出登录处理函数（纯 JavaScript，不使用 Vue）
function handleLogout() {
    if (confirm('确定要退出登录吗？')) {
        // 清除权限缓存
        if (window.PermissionManager) {
            window.PermissionManager.clearCache();
        }
        
        // 获取 token
        const token = getToken();
        
        // 调用退出登录接口
        if (token) {
            api.auth.logout().catch(function() {
                // 忽略错误，继续执行退出逻辑
            });
        }
        
        // 清除本地存储
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        
        // 清除 cookie
        document.cookie = 'token=; path=/; max-age=0';
        
        // 跳转到登录页
        window.location.href = '/login';
    }
}

/**
 * 侧边栏状态管理
 */
// 全局侧边栏状态对象
var sidebarState = {
    collapsed: getCookie('sidebarCollapsed') === 'true'
};

// 获取侧边栏初始状态（供 Vue 实例使用）
function getSidebarState() {
    return sidebarState;
}

// 更新侧边栏 UI（根据状态更新 DOM）
function updateSidebarUI() {
    var sidebar = document.getElementById('sidebar');
    var toggleIcon = document.getElementById('sidebar-toggle-icon');
    
    if (!sidebar || !toggleIcon) {
        return;
    }
    
    // 获取所有菜单项
    var menuItems = sidebar.querySelectorAll('.menu-item');
    var menuSpans = sidebar.querySelectorAll('.menu-item span');
    var logo = sidebar.querySelector('.sidebar-header .logo');
    
    if (sidebarState.collapsed) {
        // 折叠状态
        sidebar.classList.remove('w-[260px]');
        sidebar.classList.add('w-20');
        if (toggleIcon) {
            toggleIcon.className = 'pi pi-angle-right text-lg';
        }
        
        // 隐藏菜单项文字和 logo
        if (logo) {
            logo.classList.add('opacity-0', 'w-0');
        }
        menuSpans.forEach(function(span) {
            span.classList.add('opacity-0', 'w-0', 'overflow-hidden');
        });
        menuItems.forEach(function(item) {
            item.classList.add('justify-center', 'px-0');
        });
        var menuIcons = sidebar.querySelectorAll('.menu-item i');
        menuIcons.forEach(function(icon) {
            icon.classList.remove('mr-3');
        });
    } else {
        // 展开状态
        sidebar.classList.remove('w-20');
        sidebar.classList.add('w-[260px]');
        if (toggleIcon) {
            toggleIcon.className = 'pi pi-angle-left text-lg';
        }
        
        // 显示菜单项文字和 logo
        if (logo) {
            logo.classList.remove('opacity-0', 'w-0');
        }
        menuSpans.forEach(function(span) {
            span.classList.remove('opacity-0', 'w-0', 'overflow-hidden');
        });
        menuItems.forEach(function(item) {
            item.classList.remove('justify-center', 'px-0');
        });
        var menuIcons = sidebar.querySelectorAll('.menu-item i');
        menuIcons.forEach(function(icon) {
            icon.classList.add('mr-3');
        });
    }
}

// 切换侧边栏状态（纯 JavaScript 实现）
function doToggleSidebar() {
    sidebarState.collapsed = !sidebarState.collapsed;
    setCookie('sidebarCollapsed', sidebarState.collapsed, 365);
    
    // 更新 UI
    updateSidebarUI();
}

// 初始化侧边栏（页面加载完成后执行）
function initSidebar() {
    // 更新初始 UI 状态
    updateSidebarUI();
    
    // 绑定点击事件
    var toggleBtn = document.getElementById('sidebar-toggle');
    if (toggleBtn) {
        toggleBtn.addEventListener('click', doToggleSidebar);
    }
}

// DOM 加载完成后初始化侧边栏
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initSidebar);
} else {
    // DOM 已经加载完成
    initSidebar();
}

/**
 * Vue 3 基础应用配置
 * 提供所有页面共用的数据、方法和指令
 */
window.createBaseAppConfig = function() {
    return {
        data() {
            return {
                activeMenu: window.location.pathname,
                userMenuVisible: false,
                sidebarCollapsed: getCookie('sidebarCollapsed') === 'true',
                // 权限状态（用于触发响应式更新）
                permissionsReady: false,
                isSuperAdmin: false,  // 超级管理员状态
                // 菜单配置（来自 permission.js，唯一数据源）
                menuItems: window.PermissionManager && window.PermissionManager.menuItems ? window.PermissionManager.menuItems : []
            };
        },
        computed: {
            // 过滤后的菜单（根据权限）
            visibleMenus: function() {
                var self = this;
                // 如果权限管理器未初始化，至少显示不需要权限的菜单（如首页）
                if (!window.PermissionManager || !window.PermissionManager.initialized) {
                    return this.menuItems.filter(function(menu) {
                        return !menu.permission; // 只显示不需要权限的菜单
                    });
                }
                // 使用 permissionsReady 和 isSuperAdmin 触发响应式更新
                var _ = this.permissionsReady; // 读取 permissionsReady 以建立依赖关系
                var isSuperAdmin = this.isSuperAdmin; // 读取 isSuperAdmin 以建立依赖关系
                return this.menuItems.filter(function(menu) {
                    // 不需要权限的菜单始终显示
                    if (!menu.permission) {
                        return true;
                    }
                    // 超级管理员显示所有菜单（使用 Vue data 中的状态）
                    if (isSuperAdmin) {
                        return true;
                    }
                    // 检查菜单权限
                    return window.PermissionManager.isMenuVisible(menu.path);
                });
            }
        },
        methods: {
            navigate(path) {
                this.activeMenu = path;
                window.location.href = path;
            },
            closeUserMenu() {
                this.userMenuVisible = false;
            },
            toggleSidebar() {
                this.sidebarCollapsed = !this.sidebarCollapsed;
                setCookie('sidebarCollapsed', this.sidebarCollapsed, 365);
            },
            handleUserMenuCommand(command) {
                if (command === 'avatar') {
                    this.navigate('/admin/avatar');
                } else if (command === 'password') {
                    this.navigate('/admin/password');
                } else if (command === 'logout') {
                    this.handleLogout();
                }
            },
            handleDropdownVisibleChange(visible) {
                this.userMenuVisible = visible;
            },
            getUserAvatar() {
                return window.userMenuData && window.userMenuData.avatar ? window.userMenuData.avatar : '';
            },
            getUserNickname() {
                return (window.userMenuData && window.userMenuData.nickname) ? window.userMenuData.nickname : (window.userMenuData && window.userMenuData.username ? window.userMenuData.username : '');
            },
            getUserUsername() {
                return window.userMenuData && window.userMenuData.username ? window.userMenuData.username : '';
            },
            getUserRoles() {
                try {
                    var roles = window.userMenuData && window.userMenuData.roles ? window.userMenuData.roles : [];
                    var result = Array.isArray(roles) ? roles : [];
                    
                    if (typeof roles === 'string') {
                        try {
                            result = JSON.parse(roles);
                            if (!Array.isArray(result)) {
                                result = [];
                            }
                        } catch (e) {
                            result = [];
                        }
                    }
                    
                    return result;
                } catch (e) {
                    return [];
                }
            },
            getAvatar(avatar) {
                if (avatar && avatar.trim() !== '') {
                    return avatar;
                }
                return 'data:image/svg+xml;utf8,<svg width="40" height="40" viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg"><circle cx="20" cy="20" r="20" fill="%23e5e7eb"/><circle cx="20" cy="15" r="6" fill="%239ca3af"/><path d="M10 30C10 25 14 23 20 23C26 23 30 25 30 30V32H10V30Z" fill="%239ca3af"/></svg>';
            },
            handleAvatarError(event) {
                event.target.src = 'data:image/svg+xml;utf8,<svg width="40" height="40" viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg"><circle cx="20" cy="20" r="20" fill="%23e5e7eb"/><circle cx="20" cy="15" r="6" fill="%239ca3af"/><path d="M10 30C10 25 14 23 20 23C26 23 30 25 30 30V32H10V30Z" fill="%239ca3af"/></svg>';
            },
            handleLogout() {
                if (confirm('确定要退出登录吗？')) {
                    // 清除权限缓存
                    if (window.PermissionManager) {
                        window.PermissionManager.clearCache();
                    }
                    
                    api.auth.logout().then(() => {
                        localStorage.removeItem('token');
                        localStorage.removeItem('user');
                        document.cookie = 'token=; path=/; max-age=0';
                        window.location.href = '/login';
                    }).catch(() => {
                        localStorage.removeItem('token');
                        localStorage.removeItem('user');
                        document.cookie = 'token=; path=/; max-age=0';
                        window.location.href = '/login';
                    });
                }
            }
        },
        directives: {
            'click-outside': {
                beforeMount: function (el, binding) {
                    el.clickOutsideEvent = function (event) {
                        if (!(el == event.target || el.contains(event.target))) {
                            binding.instance[binding.expression](event);
                        }
                    };
                    document.body.addEventListener('click', el.clickOutsideEvent);
                },
                unmounted: function (el) {
                    document.body.removeEventListener('click', el.clickOutsideEvent);
                }
            }
        }
    };
};

/**
 * 合并 Vue 组件配置的辅助函数
 * 用于将页面特定的配置与基础配置合并
 */
window.mergeAppConfig = function(baseConfig, pageConfig) {
    // 深度合并函数
    function deepMerge(target, source) {
        var result = {};
        
        // 复制 target 的所有属性
        for (var key in target) {
            if (target.hasOwnProperty(key)) {
                result[key] = target[key];
            }
        }
        
        // 合并 source 的属性
        for (var key in source) {
            if (source.hasOwnProperty(key)) {
                if (typeof source[key] === 'object' && source[key] !== null && !Array.isArray(source[key]) && typeof source[key].constructor === 'function' && source[key].constructor === Object) {
                    // 如果是对象，递归合并
                    result[key] = deepMerge(result[key] || {}, source[key]);
                } else if (key === 'data' && typeof source[key] === 'function') {
                    // 特殊处理 data 函数
                    var baseData = typeof result[key] === 'function' ? result[key]() : {};
                    var pageData = source[key]();
                    result[key] = function() {
                        return Object.assign({}, baseData, pageData);
                    };
                } else if (key === 'methods' || key === 'computed' || key === 'directives') {
                    // 合并 methods、computed、directives
                    result[key] = Object.assign({}, result[key] || {}, source[key] || {});
                } else {
                    // 其他属性直接覆盖
                    result[key] = source[key];
                }
            }
        }
        
        return result;
    }
    
    return deepMerge(baseConfig, pageConfig);
};

/**
 * 安全挂载 Vue 应用的辅助函数
 * 如果已经有应用实例挂载，会先卸载再挂载新的实例
 */
window.safeMountApp = function(app, selector) {
    selector = selector || '#app';
    const container = document.querySelector(selector);
    if (!container) {
        return null;
    }
    
    // 检查是否已经有应用实例挂载
    if (container.__vue_app__) {
        container.__vue_app__.unmount();
    }
    
    // 挂载新应用并返回实例
    return app.mount(selector);
};

/**
 * 获取 Element Plus 组件的辅助函数
 * 确保在 Element Plus 加载后可以安全访问组件
 */
window.getElementPlusComponent = function(componentName) {
    // 尝试多种方式获取组件
    if (window[componentName]) {
        return window[componentName];
    }
    if (window.ElementPlus && window.ElementPlus[componentName]) {
        return window.ElementPlus[componentName];
    }
    if (window.ElementPlus && window.ElementPlus.ElMessage && componentName === 'ElMessage') {
        return window.ElementPlus.ElMessage;
    }
    return null;
};
