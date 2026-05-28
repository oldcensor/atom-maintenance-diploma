import React, { useState } from 'react';
import { toast } from 'react-toastify';
import { useGetDepartmentsQuery, useCreateEmployeeMutation } from '../store/apiSlice';

export default function EmployeeCreate({ onSuccess, onCancel }) {
    const { data: departments = [] } = useGetDepartmentsQuery();
    const [createEmployee, { isLoading: isPending }] = useCreateEmployeeMutation();

    const [form, setForm] = useState({ email: '', password: '', full_name: '', role: 'technician', department_id: '' });
    const set = (field) => (e) => setForm(f => ({ ...f, [field]: e.target.value }));

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await createEmployee({
                email: form.email, password: form.password, full_name: form.full_name, role: form.role,
                department_id: form.department_id ? parseInt(form.department_id, 10) : null,
            }).unwrap();
            toast.success('Сотрудник создан');
            onSuccess && onSuccess();
        } catch (err) {
            toast.error(err?.data?.error || 'Ошибка при создании');
        }
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">ФИО *</label>
                <input required maxLength={100} value={form.full_name} onChange={set('full_name')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Иванов Иван Иванович" />
            </div>
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Email *</label>
                <input required type="email" maxLength={150} value={form.email} onChange={set('email')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="ivan@aes.ru" />
            </div>
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Пароль *</label>
                <input required type="password" minLength={8} maxLength={100} value={form.password} onChange={set('password')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Минимум 8 символов" />
            </div>
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Роль *</label>
                <select value={form.role} onChange={set('role')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="technician">Техник</option><option value="engineer">Инженер</option>
                    <option value="manager">Менеджер</option><option value="admin">Администратор</option>
                </select>
            </div>
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Подразделение</label>
                <select value={form.department_id} onChange={set('department_id')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Не указано</option>
                    {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                </select>
            </div>
            <div className="flex gap-3 pt-2">
                <button type="submit" disabled={isPending} className="bg-blue-800 text-white px-6 py-2 rounded-lg text-sm hover:bg-blue-900 disabled:opacity-60 transition-colors">{isPending ? 'Создание...' : 'Создать'}</button>
                {onCancel && <button type="button" onClick={onCancel} className="px-6 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50">Отмена</button>}
            </div>
        </form>
    );
}
