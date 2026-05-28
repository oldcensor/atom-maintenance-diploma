import React, { useState } from 'react';
import { useAuth } from './AuthContext';
import { useNavigate } from 'react-router-dom';

function Login() {
    const { login } = useAuth();
    const navigate = useNavigate();

    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [errorMsg, setErrorMsg] = useState('');
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setErrorMsg('');
        setLoading(true);
        try {
            await login(email, password);
            navigate('/');
        } catch (err) {
            setErrorMsg(err.message || 'Неверный логин или пароль');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-gray-100 flex flex-col justify-center">
            <div className="text-center mb-8">
                <h1 className="text-3xl font-bold text-blue-800">ТОИР АЭС</h1>
                <p className="text-gray-500 mt-1 text-sm">Система технического обслуживания и ремонта</p>
            </div>

            <div className="mx-auto w-full max-w-md px-4">
                <div className="bg-white rounded-xl shadow-md border border-gray-200 p-8">
                    <h2 className="text-xl font-semibold text-gray-800 mb-6">Вход в систему</h2>

                    {errorMsg && (
                        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
                            {errorMsg}
                        </div>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div className="relative">
                            <input
                                id="email"
                                type="email"
                                required
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 peer"
                                placeholder=" "
                            />
                            <label
                                htmlFor="email"
                                className={`absolute left-3 transform -translate-y-1/2 text-gray-500 text-sm transition-all duration-200
                                    ${email ? '-top-[1px] text-xs bg-white px-1' : 'top-1/2'}
                                    peer-focus:-top-[1px] peer-focus:text-xs peer-focus:bg-white peer-focus:px-1`}
                            >
                                Email
                            </label>
                        </div>

                        <div className="relative">
                            <input
                                id="password"
                                type="password"
                                required
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 peer"
                                placeholder=" "
                            />
                            <label
                                htmlFor="password"
                                className={`absolute left-3 transform -translate-y-1/2 text-gray-500 text-sm transition-all duration-200
                                    ${password ? '-top-[1px] text-xs bg-white px-1' : 'top-1/2'}
                                    peer-focus:-top-[1px] peer-focus:text-xs peer-focus:bg-white peer-focus:px-1`}
                            >
                                Пароль
                            </label>
                        </div>

                        <button
                            type="submit"
                            disabled={loading}
                            className="w-full py-2 mt-2 bg-blue-800 text-white rounded-lg hover:bg-blue-900 disabled:opacity-60 transition-colors font-medium"
                        >
                            {loading ? 'Вход...' : 'Войти'}
                        </button>
                    </form>
                </div>
            </div>
        </div>
    );
}

export default Login;
