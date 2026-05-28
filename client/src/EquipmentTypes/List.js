import React, { useState } from 'react';
import Loading from '../utils/Loading';
import { toast } from 'react-toastify';
import { useGetEquipmentTypesQuery, useCreateEquipmentTypeMutation, useUpdateEquipmentTypeMutation, useDeleteEquipmentTypeMutation } from '../store/apiSlice';
import { useAuth, hasAtLeastRole } from '../auth/AuthContext';

export default function EquipmentTypeList() {
    const { user } = useAuth();
    const canEdit = hasAtLeastRole(user?.role, 'admin');
    const { data: types = [], isLoading: isPending, error } = useGetEquipmentTypesQuery();
    const [createType, { isLoading: creating }] = useCreateEquipmentTypeMutation();
    const [updateType, { isLoading: updating }] = useUpdateEquipmentTypeMutation();
    const [deleteType] = useDeleteEquipmentTypeMutation();

    const [showForm, setShowForm] = useState(false);
    const [editId, setEditId] = useState(null);
    const [form, setForm] = useState({ name: '', description: '' });
    const set = (field) => (e) => setForm(f => ({ ...f, [field]: e.target.value }));

    const startEdit = (t) => { setEditId(t.id); setForm({ name: t.name, description: t.description || '' }); setShowForm(true); };
    const startCreate = () => { setEditId(null); setForm({ name: '', description: '' }); setShowForm(true); };

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            if (editId) {
                await updateType({ id: editId, name: form.name, description: form.description }).unwrap();
                toast.success('Тип обновлён');
            } else {
                await createType({ name: form.name, description: form.description }).unwrap();
                toast.success('Тип создан');
            }
            setShowForm(false);
        } catch (err) {
            toast.error(err?.data?.error || 'Ошибка');
        }
    };

    const handleDelete = async (t) => {
        if (!window.confirm(`Удалить тип "${t.name}"?`)) return;
        try {
            await deleteType(t.id).unwrap();
            toast.success('Удалено');
        } catch (err) {
            toast.error(err?.data?.error || 'Ошибка');
        }
    };

    return (
        <div className="space-y-4 max-w-2xl">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-800">Типы оборудования</h1>
                {canEdit && <button onClick={startCreate} className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900 transition-colors">+ Создать</button>}
            </div>
            {canEdit && showForm && (
                <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
                    <h3 className="font-semibold text-gray-800 mb-4">{editId ? 'Редактировать' : 'Новый тип'}</h3>
                    <form onSubmit={handleSubmit} className="space-y-3">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Название *</label>
                            <input required maxLength={100} value={form.name} onChange={set('name')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Насос главный циркуляционный" />
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
                : types.length === 0 ? <div className="px-5 py-12 text-center text-gray-400">Нет типов оборудования</div>
                : (
                    <table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase"><tr>
                            <th className="px-5 py-3 text-left">ID</th><th className="px-5 py-3 text-left">Название</th>
                            <th className="px-5 py-3 text-left">Описание</th>
                            {canEdit && <th className="px-5 py-3 text-left">Действия</th>}
                        </tr></thead>
                        <tbody className="divide-y divide-gray-100">{types.map(t => (
                            <tr key={t.id} className="hover:bg-gray-50">
                                <td className="px-5 py-3 text-gray-400">#{t.id}</td>
                                <td className="px-5 py-3 font-medium text-gray-800">{t.name}</td>
                                <td className="px-5 py-3 text-gray-500">{t.description || '—'}</td>
                                {canEdit && <td className="px-5 py-3"><div className="flex gap-3">
                                    <button onClick={() => startEdit(t)} className="text-xs text-blue-600 hover:text-blue-800">Изменить</button>
                                    <button onClick={() => handleDelete(t)} className="text-xs text-red-500 hover:text-red-700">Удалить</button>
                                </div></td>}
                            </tr>
                        ))}</tbody>
                    </table>
                )}
            </div>
        </div>
    );
}
