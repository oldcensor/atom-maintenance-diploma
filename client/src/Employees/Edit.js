import React, { useState } from 'react';
import { toast } from 'react-toastify';
import { useGetDepartmentsQuery, useUpdateEmployeeMutation } from '../store/apiSlice';

export default function EmployeeEdit({ employee, onSuccess, onCancel }) {
    const { data: departments = [] } = useGetDepartmentsQuery();
    const [updateEmployee, { isLoading: isPending }] = useUpdateEmployeeMutation();

    const [form, setForm] = useState({
        full_name: employee.full_name || '',
        email: employee.email || '',
        role: employee.role || 'technician',
        department_id: employee.department_id ? String(employee.department_id) : '',
    });
    const set = (field) => (e) => setForm(f => ({ ...f, [field]: e.target.value }));

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await updateEmployee({
                id: employee.id,
                full_name: form.full_name,
                email: form.email,
                role: form.role,
                department_id: form.department_id ? parseInt(form.department_id, 10) : null,
            }).unwrap();
            toast.success('Данные сотрудника обновлены');
            onSuccess && onSuccess();
        } catch (err) {
            toast.error(err?.data?.error || 'Ошибка при обновлении');
        }
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">ФИО *</label>
                <input required value={form.full_name} onChange={set('full_name')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Иванов Иван Иванович" />
            </div>
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Email *</label>
                <input required type="email" value={form.email} onChange={set('email')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Роль *</label>
                <select value={form.role} onChange={set('role')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="technician">Техник</option>
                    <option value="engineer">Инженер</option>
                    <option value="manager">Менеджер</option>
                    <option value="admin">Администратор</option>
                </select>
            </div>
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Подразделение</label>
                <select value={form.department_id} onChange={set('department_id')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Не указано</option>
                    {departments.map(d => <option key={d.id} value={String(d.id)}>{d.name}</option>)}
                </select>
            </div>
            <div className="flex gap-3 pt-2">
                <button type="submit" disabled={isPending} className="bg-blue-800 text-white px-6 py-2 rounded-lg text-sm hover:bg-blue-900 disabled:opacity-60 transition-colors">
                    {isPending ? 'Сохранение...' : 'Сохранить'}
                </button>
                {onCancel && <button type="button" onClick={onCancel} className="px-6 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50">Отмена</button>}
            </div>
        </form>
    );
}
