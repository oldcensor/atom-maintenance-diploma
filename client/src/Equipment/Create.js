import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'react-toastify';
import { useGetEquipmentTypesQuery, useGetDepartmentsQuery, useGetEmployeesQuery, useCreateEquipmentMutation } from '../store/apiSlice';

export default function EquipmentCreate() {
    const navigate = useNavigate();
    const { data: types = [] } = useGetEquipmentTypesQuery();
    const { data: departments = [] } = useGetDepartmentsQuery();
    const { data: employees = [] } = useGetEmployeesQuery();
    const [createEquipment, { isLoading: isPending }] = useCreateEquipmentMutation();

    const [form, setForm] = useState({ name: '', description: '', serial_number: '', equipment_type_id: '', department_id: '', responsible_id: '', status: 'active' });
    const set = (field) => (e) => setForm(f => ({ ...f, [field]: e.target.value }));

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await createEquipment({
                name: form.name, description: form.description, serial_number: form.serial_number,
                equipment_type_id: parseInt(form.equipment_type_id, 10),
                department_id: form.department_id ? parseInt(form.department_id, 10) : null,
                responsible_id: form.responsible_id ? parseInt(form.responsible_id, 10) : null,
                status: form.status,
            }).unwrap();
            toast.success('Оборудование создано');
            navigate('/equipment');
        } catch (err) {
            toast.error(err?.data?.error || 'Ошибка при создании');
        }
    };

    return (
        <div className="max-w-2xl">
            <div className="flex items-center gap-3 mb-6">
                <button onClick={() => navigate('/equipment')} className="text-gray-400 hover:text-gray-600">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5"><path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" /></svg>
                </button>
                <h1 className="text-2xl font-bold text-gray-800">Новое оборудование</h1>
            </div>
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="col-span-2">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Название *</label>
                            <input required maxLength={100} value={form.name} onChange={set('name')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="ГЦН-1" />
                        </div>
                        <div className="col-span-2">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Серийный номер *</label>
                            <input required maxLength={50} value={form.serial_number} onChange={set('serial_number')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="SN-00001" />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Тип оборудования *</label>
                            <select required value={form.equipment_type_id} onChange={set('equipment_type_id')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                <option value="">Выберите тип</option>
                                {types.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
                            </select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Подразделение</label>
                            <select value={form.department_id} onChange={set('department_id')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                <option value="">Не указано</option>
                                {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                            </select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Ответственный</label>
                            <select value={form.responsible_id} onChange={set('responsible_id')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                <option value="">Не назначен</option>
                                {employees.map(e => <option key={e.id} value={e.id}>{e.full_name}</option>)}
                            </select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Статус</label>
                            <select value={form.status} onChange={set('status')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                <option value="active">Активно</option>
                                <option value="inactive">Неактивно</option>
                                <option value="under_maintenance">На ремонте</option>
                            </select>
                        </div>
                        <div className="col-span-2">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
                            <textarea rows={3} maxLength={1000} value={form.description} onChange={set('description')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none" placeholder="Дополнительная информация об оборудовании" />
                        </div>
                    </div>
                    <div className="flex gap-3 pt-2">
                        <button type="submit" disabled={isPending} className="bg-blue-800 text-white px-6 py-2 rounded-lg text-sm hover:bg-blue-900 disabled:opacity-60 transition-colors">{isPending ? 'Создание...' : 'Создать'}</button>
                        <button type="button" onClick={() => navigate('/equipment')} className="px-6 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50 transition-colors">Отмена</button>
                    </div>
                </form>
            </div>
        </div>
    );
}
