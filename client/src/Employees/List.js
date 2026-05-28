import React, { useState } from 'react';
import { useAuth, hasAtLeastRole } from '../auth/AuthContext';
import Loading from '../utils/Loading';
import { toast } from 'react-toastify';
import EmployeeCreate from './Create';
import EmployeeEdit from './Edit';
import { useGetEmployeesQuery, useGetDepartmentsQuery, useDeleteEmployeeMutation } from '../store/apiSlice';

const ROLE_LABELS = { technician: 'Техник', engineer: 'Инженер', manager: 'Менеджер', admin: 'Администратор' };
const ROLE_COLORS = { technician: 'bg-gray-100 text-gray-600', engineer: 'bg-blue-100 text-blue-700', manager: 'bg-purple-100 text-purple-700', admin: 'bg-red-100 text-red-700' };

export default function EmployeeList() {
    const { user } = useAuth();
    const { data: employees = [], isLoading: isPending, error } = useGetEmployeesQuery();
    const { data: departments = [] } = useGetDepartmentsQuery();
    const [deleteEmployee] = useDeleteEmployeeMutation();

    const [filterRole, setFilterRole] = useState('');
    const [filterDept, setFilterDept] = useState('');
    const [showCreate, setShowCreate] = useState(false);
    const [editEmployee, setEditEmployee] = useState(null);

    const deptMap = Object.fromEntries(departments.map(d => [d.id, d.name]));
    const filtered = employees.filter(e => {
        if (filterRole && e.role !== filterRole) return false;
        if (filterDept && String(e.department_id) !== filterDept) return false;
        return true;
    });

    const handleDelete = async (emp) => {
        if (!window.confirm(`Удалить сотрудника ${emp.full_name}?`)) return;
        try {
            await deleteEmployee(emp.id).unwrap();
            toast.success('Сотрудник удалён');
        } catch (err) {
            toast.error(err?.data?.error || 'Ошибка');
        }
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-800">Сотрудники</h1>
                {hasAtLeastRole(user?.role, 'admin') && (
                    <button onClick={() => setShowCreate(true)} className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900 transition-colors">+ Создать</button>
                )}
            </div>
            <div className="flex flex-wrap gap-3">
                <select value={filterRole} onChange={e => setFilterRole(e.target.value)} className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Все роли</option>
                    {Object.entries(ROLE_LABELS).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
                </select>
                <select value={filterDept} onChange={e => setFilterDept(e.target.value)} className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Все подразделения</option>
                    {departments.map(d => <option key={d.id} value={String(d.id)}>{d.name}</option>)}
                </select>
                {(filterRole || filterDept) && (
                    <button onClick={() => { setFilterRole(''); setFilterDept(''); }} className="text-sm text-gray-500 hover:text-gray-700 px-2">Сбросить</button>
                )}
            </div>
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                {isPending ? <div className="flex justify-center py-12"><Loading /></div>
                : error ? <div className="px-5 py-8 text-center text-red-500">{error?.data?.error || 'Ошибка загрузки'}</div>
                : filtered.length === 0 ? <div className="px-5 py-12 text-center text-gray-400">Нет данных</div>
                : (
                    <div className="overflow-x-auto"><table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase"><tr>
                            <th className="px-5 py-3 text-left">ID</th><th className="px-5 py-3 text-left">ФИО</th>
                            <th className="px-5 py-3 text-left">Email</th><th className="px-5 py-3 text-left">Роль</th>
                            <th className="px-5 py-3 text-left">Подразделение</th>
                            {hasAtLeastRole(user?.role, 'manager') && <th className="px-5 py-3 text-left">Действия</th>}
                        </tr></thead>
                        <tbody className="divide-y divide-gray-100">{filtered.map(emp => (
                            <tr key={emp.id} className="hover:bg-gray-50">
                                <td className="px-5 py-3 text-gray-400">#{emp.id}</td>
                                <td className="px-5 py-3 font-medium text-gray-800">{emp.full_name}</td>
                                <td className="px-5 py-3 text-gray-500">{emp.email}</td>
                                <td className="px-5 py-3"><span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${ROLE_COLORS[emp.role] || 'bg-gray-100'}`}>{ROLE_LABELS[emp.role] || emp.role}</span></td>
                                <td className="px-5 py-3 text-gray-600">{emp.department_id ? deptMap[emp.department_id] || `#${emp.department_id}` : '—'}</td>
                                {hasAtLeastRole(user?.role, 'manager') && (
                                    <td className="px-5 py-3 flex gap-3">
                                        {hasAtLeastRole(user?.role, 'admin') && (
                                            <button onClick={() => setEditEmployee(emp)} className="text-xs text-blue-600 hover:text-blue-800">Редактировать</button>
                                        )}
                                        {hasAtLeastRole(user?.role, 'admin') && emp.id !== user?.id && (
                                            <button onClick={() => handleDelete(emp)} className="text-xs text-red-500 hover:text-red-700">Удалить</button>
                                        )}
                                    </td>
                                )}
                            </tr>
                        ))}</tbody>
                    </table></div>
                )}
            </div>

            {showCreate && (
                <div className="fixed inset-0 bg-black bg-opacity-40 flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-xl shadow-xl w-full max-w-md">
                        <div className="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
                            <h3 className="font-semibold text-gray-800">Новый сотрудник</h3>
                            <button onClick={() => setShowCreate(false)} className="text-gray-400 hover:text-gray-600">✕</button>
                        </div>
                        <div className="p-6">
                            <EmployeeCreate onSuccess={() => setShowCreate(false)} onCancel={() => setShowCreate(false)} />
                        </div>
                    </div>
                </div>
            )}

            {editEmployee && (
                <div className="fixed inset-0 bg-black bg-opacity-40 flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-xl shadow-xl w-full max-w-md">
                        <div className="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
                            <h3 className="font-semibold text-gray-800">Редактирование сотрудника</h3>
                            <button onClick={() => setEditEmployee(null)} className="text-gray-400 hover:text-gray-600">✕</button>
                        </div>
                        <div className="p-6">
                            <EmployeeEdit employee={editEmployee} onSuccess={() => setEditEmployee(null)} onCancel={() => setEditEmployee(null)} />
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
