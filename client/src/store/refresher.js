// Единственный источник правды для refresh токена.
// AuthContext регистрирует свою функцию refreshTokens(),
// apiSlice вызывает её через callRefresh() — это исключает
// двойной параллельный refresh и выброс на логин.
let _refreshFn = null;

export const registerRefresh = (fn) => { _refreshFn = fn; };
export const callRefresh = () =>
    _refreshFn ? _refreshFn() : Promise.reject(new Error('refresher not registered'));
