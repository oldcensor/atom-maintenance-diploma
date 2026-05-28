import React, { useState } from 'react';
import Loading from '../utils/Loading';
import { toast } from 'react-toastify';
import { useGetDepartmentsQuery, useCreateDepartmentMutation, useUpdateDepartmentMutation, useDeleteDepartmentMutation } from '../store/apiSlice';
import { useAuth, hasAtLeastRole } from '../auth/AuthContext';

export default function DepartmentList() {
    const { user } = useAuth();
    const canEdit = hasAtLeastRole(user?.role, 'admin');
    const { data: departments = [], isLoading: isPending, error } = useGetDepartmentsQuery();
    const [createDept, { isLoading: creating }] = useCreateDepartmentMutation();
    const [updateDept, { isLoading: updating }] = useUpdateDepartmentMutation();
    const [deleteDept] = useDeleteDepartmentMutation();

    const [showForm, setShowForm] = useState(false);
    const [editId, setEditId] = useState(null);
    const [form, setForm] = useState({ name: '', description: '' });
    const set = (field) => (e) => setForm(f => ({ ...f, [field]: e.target.value }));

    const startEdit = (dept) => { setEditId(dept.id); setForm({ name: dept.name, description: dept.description || '' }); setShowForm(true); };
    const startCreate = () => { setEditId(null); setForm({ name: '', description: '' }); setShowForm(true); };

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            if (editId) {
                await updateDept({ id: editId, name: form.name, description: form.description }).unwrap();
                toast.success('Подразделение обновлено');
            } else {
                await createDept({ name: form.name, description: form.description }).unwrap();
                toast.success('Подразделение создано');
            }
            setShowForm(false);
        } catch (err) {
            toast.error(err?.data?.error || 'Ошибка');
        }
    };

    const handleDelete = async (dept) => {
        if (!window.confirm(`Удалить подразделение "${dept.name}"?`)) return;
        try {
            await deleteDept(dept.id).unwrap();
            toast.success('Удалено');
        } catch (err) {
            toast.error(err?.data?.error || 'Ошибка');
        }
    };

    return (
        <div className="space-y-4 max-w-2xl">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-800">Подразделения</h1>
                {canEdit && <button onClick={startCreate} className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900 transition-colors">+ Создать</button>}
            </div>
            {canEdit && showForm && (
                <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
                    <h3 className="font-semibold text-gray-800 mb-4">{editId ? 'Редактировать' : 'Новое подразделение'}</h3>
                    <form onSubmit={handleSubmit} className="space-y-3">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Название *</label>
                            <input required maxLength={100} value={form.name} onChange={set('name')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Реакторный цех" />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
                            <textarea rows={2} maxLength={500} value={form.description} onChange={set('description')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none" />
                        </div>
                        <div className="flex gap-3">
                            <button type="submit" disabled={creating || updating} className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900 disabled:opacity-60">{creating || updating ? 'Сохранение...' : 'Сохранить'}</button>
                            <button type="button" onClick={() => setShowForm(false)} className="px-4 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50">Отмена</button>
                        </div>
                    </form>
                </div>
            )}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                {isPending ? <div className="flex justify-center py-12"><Loading /></div>
                : error ? <div className="px-5 py-8 text-center text-red-500">{error?.data?.error || 'Ошибка загрузки'}</div>
                : departments.length === 0 ? <div className="px-5 py-12 text-center text-gray-400">Нет подразделений</div>
                : (
                    <table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase"><tr>
                            <th className="px-5 py-3 text-left">ID</th><th className="px-5 py-3 text-left">Название</th>
                            <th className="px-5 py-3 text-left">Описание</th>
                            {canEdit && <th className="px-5 py-3 text-left">Действия</th>}
                        </tr></thead>
                        <tbody className="divide-y divide-gray-100">{departments.map(d => (
                            <tr key={d.id} className="hover:bg-gray-50">
                                <td className="px-5 py-3 text-gray-400">#{d.id}</td>
                                <td className="px-5 py-3 font-medium text-gray-800">{d.name}</td>
                                <td className="px-5 py-3 text-gray-500">{d.description || '—'}</td>
                                {canEdit && <td className="px-5 py-3"><div className="flex gap-3">
                                    <button onClick={() => startEdit(d)} className="text-xs text-blue-600 hover:text-blue-800">Изменить</button>
                                    <button onClick={() => handleDelete(d)} className="text-xs text-red-500 hover:text-red-700">Удалить</button>
                                </div></td>}
                            </tr>
                        ))}</tbody>
                    </table>
                )}
            </div>
        </div>
    );
}
