import { useState } from 'react';
import { useAuth, hasAtLeastRole } from './auth/AuthContext';
import { useNavigate, NavLink } from 'react-router-dom';

const ROLE_LABELS = {
    technician: 'Техник',
    engineer: 'Инженер',
    manager: 'Менеджер',
    admin: 'Администратор',
};

function Navbar() {
    const { logout, user } = useAuth();
    const navigate = useNavigate();
    const [isDropdownOpen, setIsDropdownOpen] = useState(false);
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

    const handleLogout = async (e) => {
        e.preventDefault();
        await logout();
        navigate('/login');
    };

    const navLinkClass = ({ isActive }) =>
        `px-3 py-2 rounded-md text-sm font-medium transition-colors ${
            isActive
                ? 'bg-blue-900 text-white'
                : 'text-blue-100 hover:bg-blue-700 hover:text-white'
        }`;

    return (
        <nav className="bg-blue-800 shadow-md">
            <div className="max-w-screen-xl mx-auto px-4">
                <div className="flex items-center h-14">
                    {/* Логотип */}
                    <div className="flex-shrink-0 flex items-center mr-6">
                        <span className="text-white font-bold text-lg tracking-tight">ТОИР АЭС</span>
                    </div>

                    {/* Основная навигация */}
                    <div className="hidden md:flex items-center gap-1 flex-1">
                        <NavLink to="/" end className={navLinkClass}>Сводка</NavLink>
                        <NavLink to="/equipment" className={navLinkClass}>Оборудование</NavLink>
                        <NavLink to="/schedules" className={navLinkClass}>Регламенты</NavLink>
                        <NavLink to="/work-orders" className={navLinkClass}>Наряды</NavLink>
                    </div>

                    {/* Профиль */}
                    <div
                        className="relative ml-auto"
                        onMouseEnter={() => setIsDropdownOpen(true)}
                        onMouseLeave={() => setIsDropdownOpen(false)}
                    >
                        <button className="flex items-center gap-2 text-blue-100 hover:text-white hover:bg-blue-700 rounded-md px-3 py-2 transition-colors text-sm">
                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                                <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" />
                            </svg>
                            <span className="hidden md:inline max-w-[140px] truncate">{user?.full_name || 'Пользователь'}</span>
                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                <path strokeLinecap="round" strokeLinejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5" />
                            </svg>
                        </button>

                        {isDropdownOpen && (
                            <div className="absolute right-0 top-full w-56 bg-white border border-gray-200 shadow-lg rounded-md z-50">
                                <div className="px-4 py-3 border-b border-gray-100">
                                    <p className="text-sm font-semibold text-gray-800 truncate">{user?.full_name}</p>
                                    <p className="text-xs text-gray-500 truncate">{user?.email}</p>
                                    <p className="text-xs text-blue-600 font-medium mt-1">{ROLE_LABELS[user?.role] || user?.role}</p>
                                </div>

                                <div className="py-1">
                                    {user && hasAtLeastRole(user.role, 'manager') && (
                                        <button
                                            onClick={() => { navigate('/employees'); setIsDropdownOpen(false); }}
                                            className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2"
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                                <path strokeLinecap="round" strokeLinejoin="round" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" />
                                            </svg>
                                            Сотрудники
                                        </button>
                                    )}
                                    {user && hasAtLeastRole(user.role, 'admin') && (
                                        <>
                                            <button
                                                onClick={() => { navigate('/departments'); setIsDropdownOpen(false); }}
                                                className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2"
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                                    <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />
                                                </svg>
                                                Подразделения
                                            </button>
                                            <button
                                                onClick={() => { navigate('/equipment-types'); setIsDropdownOpen(false); }}
                                                className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2"
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                                    <path strokeLinecap="round" strokeLinejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z" />
                                                    <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
                                                </svg>
                                                Типы оборудования
                                            </button>
                                        </>
                                    )}
                                </div>

                                <div className="border-t border-gray-100 py-1">
                                    <button
                                        onClick={handleLogout}
                                        className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                            <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 9V5.25A2.25 2.25 0 0 1 10.5 3h6a2.25 2.25 0 0 1 2.25 2.25v13.5A2.25 2.25 0 0 1 16.5 21h-6a2.25 2.25 0 0 1-2.25-2.25V15m-3 0-3-3m0 0 3-3m-3 3H15" />
                                        </svg>
                                        Выйти
                                    </button>
                                </div>
                            </div>
                        )}
                    </div>

                    {/* Кнопка мобильного меню */}
                    <button
                        className="md:hidden ml-2 text-blue-100 hover:text-white p-2"
                        onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
                        </svg>
                    </button>
                </div>

                {isMobileMenuOpen && (
                    <div className="md:hidden pb-3 border-t border-blue-700 pt-2 flex flex-col gap-1">
                        <NavLink to="/" end className={navLinkClass} onClick={() => setIsMobileMenuOpen(false)}>Сводка</NavLink>
                        <NavLink to="/equipment" className={navLinkClass} onClick={() => setIsMobileMenuOpen(false)}>Оборудование</NavLink>
                        <NavLink to="/schedules" className={navLinkClass} onClick={() => setIsMobileMenuOpen(false)}>Регламенты</NavLink>
                        <NavLink to="/work-orders" className={navLinkClass} onClick={() => setIsMobileMenuOpen(false)}>Наряды</NavLink>
                        {user && hasAtLeastRole(user.role, 'manager') && (
                            <NavLink to="/employees" className={navLinkClass} onClick={() => setIsMobileMenuOpen(false)}>Сотрудники</NavLink>
                        )}
                    </div>
                )}
            </div>
        </nav>
    );
}

export default Navbar;
