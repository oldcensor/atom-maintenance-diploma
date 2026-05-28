import React, { createContext, useState, useEffect, useContext, useRef, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import { tokenSet, tokenCleared } from '../store/authSlice';
import { registerRefresh } from '../store/refresher';
import Config from '../utils/Config';

const AuthContext = createContext();
const BASE = Config.endpoints.baseUrl;

export const ROLE_ORDER = ['technician', 'engineer', 'manager', 'admin'];

export const hasAtLeastRole = (userRole, minRole) => {
    return ROLE_ORDER.indexOf(userRole) >= ROLE_ORDER.indexOf(minRole);
};

function decodeJWT(token) {
    try {
        const b64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
        return JSON.parse(atob(b64));
    } catch {
        return null;
    }
}

export const AuthProvider = ({ children }) => {
    const dispatch = useDispatch();
    const [accessToken, setAccessToken] = useState(null);
    const [user, setUser] = useState(null);
    const [loadingAuth, setLoadingAuth] = useState(true);

    const tokenRef = useRef(null);
    const isRefreshingRef = useRef(false);
    const refreshQueueRef = useRef([]);

    const setToken = (t) => {
        tokenRef.current = t;
        setAccessToken(t);
        if (t) dispatch(tokenSet(t)); else dispatch(tokenCleared());
    };

    const getToken = useCallback(() => tokenRef.current, []);

    const fetchEmployee = async (token, id) => {
        const res = await fetch(`${BASE}/api/v1/employees/${id}`, {
            headers: { Authorization: `Bearer ${token}` },
        });
        if (!res.ok) return null;
        return res.json();
    };

    const performRefresh = useCallback(async () => {
        const stored = localStorage.getItem('refreshToken');
        if (!stored) throw new Error('no refresh token');
        const res = await fetch(`${BASE}/api/v1/auth/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: stored }),
        });
        if (!res.ok) throw new Error('refresh failed');
        const data = await res.json();
        localStorage.setItem('refreshToken', data.refresh_token);
        setToken(data.access_token);
        return data.access_token;
    }, []);

    const refreshTokens = useCallback(async () => {
        if (isRefreshingRef.current) {
            return new Promise((resolve, reject) => {
                refreshQueueRef.current.push({ resolve, reject });
            });
        }
        isRefreshingRef.current = true;
        try {
            const newToken = await performRefresh();
            refreshQueueRef.current.forEach(({ resolve }) => resolve(newToken));
            refreshQueueRef.current = [];
            return newToken;
        } catch (err) {
            refreshQueueRef.current.forEach(({ reject }) => reject(err));
            refreshQueueRef.current = [];
            throw err;
        } finally {
            isRefreshingRef.current = false;
        }
    }, [performRefresh]);

    useEffect(() => {
        registerRefresh(refreshTokens);
    }, [refreshTokens]);

    useEffect(() => {
        const stored = localStorage.getItem('refreshToken');
        if (!stored) { setLoadingAuth(false); return; }

        refreshTokens()
            .then(async (newToken) => {
                const payload = decodeJWT(newToken);
                if (!payload) return;
                const emp = await fetchEmployee(newToken, parseInt(payload.sub, 10));
                setUser(emp
                    ? { ...emp, role: payload.role }
                    : { id: parseInt(payload.sub, 10), role: payload.role });
            })
            .catch(() => localStorage.removeItem('refreshToken'))
            .finally(() => setLoadingAuth(false));
    }, []); // eslint-disable-line react-hooks/exhaustive-deps

    const login = async (email, password) => {
        const res = await fetch(`${BASE}/api/v1/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password }),
        });
        if (!res.ok) {
            const body = await res.json().catch(() => ({}));
            throw new Error(body.error || body.message || 'Неверный логин или пароль');
        }
        const data = await res.json();
        localStorage.setItem('refreshToken', data.refresh_token);
        setToken(data.access_token);
        const payload = decodeJWT(data.access_token);
        const emp = await fetchEmployee(data.access_token, parseInt(payload.sub, 10));
        setUser(emp
            ? { ...emp, role: payload.role }
            : { id: parseInt(payload.sub, 10), role: payload.role });
    };

    const logout = useCallback(async () => {
        try {
            if (tokenRef.current) {
                await fetch(`${BASE}/api/v1/auth/logout`, {
                    method: 'POST',
                    headers: { Authorization: `Bearer ${tokenRef.current}` },
                });
            }
        } catch { /* ignore */ } finally {
            localStorage.removeItem('refreshToken');
            setToken(null);
            setUser(null);
        }
    }, []);

    return (
        <AuthContext.Provider value={{
            accessToken, user, loadingAuth,
            isAuthenticated: !!accessToken && !!user,
            login, logout, getToken, refreshTokens,
        }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => useContext(AuthContext);
