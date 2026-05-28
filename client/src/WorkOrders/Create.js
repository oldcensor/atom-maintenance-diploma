import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'react-toastify';
import { useGetEquipmentQuery, useGetEmployeesQuery, useGetMaintenanceSchedulesQuery, useCreateWorkOrderMutation, useUpdateEquipmentMutation } from '../store/apiSlice';
import { useAuth } from '../auth/AuthContext';

const ROLE_LABELS = { technician: 'Техник', engineer: 'Инженер', manager: 'Менеджер', admin: 'Администратор' };
const ASSIGNABLE_ROLES = ['technician', 'engineer'];

const WO_TYPES = {
    emergency:  { label: 'Аварийный',      description: 'Немедленное устранение неисправности', activeClass: 'border-red-500 bg-red-50 text-red-700', dotClass: 'bg-red-500' },
    corrective: { label: 'Корректирующий', description: 'Устранение выявленной проблемы',          activeClass: 'border-yellow-500 bg-yellow-50 text-yellow-700', dotClass: 'bg-yellow-500' },
    planned:    { label: 'Плановый',        description: 'Техническое обслуживание по регламенту',  activeClass: 'border-blue-500 bg-blue-50 text-blue-700', dotClass: 'bg-blue-500' },
};

export default function WorkOrderCreate() {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();
    const { user } = useAuth();
    const { data: equipment = [] } = useGetEquipmentQuery();
    const { data: employees = [] } = useGetEmployeesQuery();
    const { data: schedules = [] } = useGetMaintenanceSchedulesQuery();
    const [createWorkOrder, { isLoading: isPending }] = useCreateWorkOrderMutation();
    const [updateEquipment] = useUpdateEquipmentMutation();

    const [woType, setWoType] = useState('corrective');
    const [form, setForm] = useState({ equipment_id: searchParams.get('equipment_id') || '', title: '', description: '', assigned_to: '', schedule_id: '', status: 'open' });
    const set = (field) => (e) => setForm(f => ({ ...f, [field]: e.target.value }));

    // Авто-подстановка ответственного из оборудования при смене оборудования
    useEffect(() => {
        if (!form.equipment_id) return;
        const eq = equipment.find(e => e.id === parseInt(form.equipment_id, 10));
        setForm(f => ({ ...f, assigned_to: eq?.responsible_id ? String(eq.responsible_id) : '' }));
    }, [form.equipment_id, equipment]); // eslint-disable-line react-hooks/exhaustive-deps

    const relevantSchedules = form.equipment_id
        ? schedules.filter(s => String(s.equipment_id) === form.equipment_id && s.status !== 'cancelled' && s.status !== 'completed')
        : [];

    const handleTypeChange = (type) => {
        setWoType(type);
        if (type !== 'planned') setForm(f => ({ ...f, schedule_id: '' }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await createWorkOrder({
                equipment_id: parseInt(form.equipment_id, 10),
                title: form.title,
                description: form.description,
                assigned_to: form.assigned_to ? parseInt(form.assigned_to, 10) : null,
                schedule_id: (woType === 'planned' && form.schedule_id) ? parseInt(form.schedule_id, 10) : null,
                status: form.status,
                work_type: woType,
            }).unwrap();

            if (form.status === 'in_progress') {
                const eq = equipment.find(e => e.id === parseInt(form.equipment_id, 10));
                if (eq) {
                    try {
                        await updateEquipment({
                            id: eq.id,
                            name: eq.name,
                            description: eq.description || '',
                            serial_number: eq.serial_number,
                            equipment_type_id: eq.equipment_type_id,
                            department_id: eq.department_id || null,
                            responsible_id: eq.responsible_id || null,
                            status: 'under_maintenance',
                        }).unwrap();
                    } catch {
                        // нет прав на обновление оборудования — не блокируем создание наряда
                    }
                }
            }

            toast.success('Наряд создан');
            navigate('/work-orders');
        } catch (err) {
            toast.error(err?.data?.error || 'Ошибка при создании');
        }
    };

    return (
        <div className="max-w-2xl">
            <div className="flex items-center gap-3 mb-6">
                <button onClick={() => navigate('/work-orders')} className="text-gray-400 hover:text-gray-600">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5"><path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" /></svg>
                </button>
                <h1 className="text-2xl font-bold text-gray-800">Новый наряд</h1>
            </div>

            {/* Тип наряда */}
            <div className="mb-4">
                <p className="text-sm font-medium text-gray-700 mb-2">Тип наряда *</p>
                <div className="grid grid-cols-3 gap-3">
                    {Object.entries(WO_TYPES).map(([key, t]) => (
                        <button
                            key={key}
                            type="button"
                            onClick={() => handleTypeChange(key)}
                            className={`flex flex-col items-start p-3 rounded-xl border-2 text-left transition-all ${woType === key ? t.activeClass : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300'}`}
                        >
                            <div className="flex items-center gap-2 mb-1">
                                <span className={`w-2 h-2 rounded-full ${t.dotClass}`} />
                                <span className="font-medium text-sm">{t.label}</span>
                            </div>
                            <span className="text-xs opacity-75">{t.description}</span>
                        </button>
                    ))}
                </div>
            </div>

            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="col-span-2">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Заголовок *</label>
                            <input required maxLength={200} value={form.title} onChange={set('title')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Краткое описание работ" />
                        </div>
                        <div className="col-span-2">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Оборудование *</label>
                            <select required value={form.equipment_id} onChange={e => { set('equipment_id')(e); setForm(f => ({ ...f, schedule_id: '' })); }} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                <option value="">Выберите оборудование</option>
                                {equipment.filter(e => e.status !== 'decommissioned').map(e => <option key={e.id} value={e.id}>{e.name} ({e.serial_number})</option>)}
                            </select>
                        </div>
                        {woType === 'planned' && relevantSchedules.length > 0 && (
                            <div className="col-span-2">
                                <label className="block text-sm font-medium text-gray-700 mb-1">Регламент</label>
                                <select value={form.schedule_id} onChange={set('schedule_id')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                    <option value="">Без привязки к регламенту</option>
                                    {relevantSchedules.map(s => <option key={s.id} value={s.id}>Регламент #{s.id} — {s.description || `${s.interval_value} ${s.interval_unit}`}</option>)}
                                </select>
                            </div>
                        )}
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Исполнитель</label>
                            <select value={form.assigned_to} onChange={set('assigned_to')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                <option value="">Не назначен</option>
                                {employees.filter(emp => ASSIGNABLE_ROLES.includes(emp.role)).map(emp => <option key={emp.id} value={emp.id}>{emp.full_name} ({ROLE_LABELS[emp.role] ?? emp.role})</option>)}
                            </select>
                            {form.assigned_to && equipment.find(e => e.id === parseInt(form.equipment_id, 10))?.responsible_id === parseInt(form.assigned_to, 10) && (
                                <p className="mt-1 text-xs text-blue-500">Автоматически назначен ответственный за это оборудование</p>
                            )}
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Начальный статус</label>
                            <select value={form.status} onChange={set('status')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                <option value="open">Открыт</option>
                                <option value="in_progress">В работе</option>
                            </select>
                        </div>
                        <div className="col-span-2">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
                            <textarea rows={3} maxLength={1000} value={form.description} onChange={set('description')} className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none" placeholder="Подробное описание работ" />
                        </div>
                    </div>
                    <div className="flex gap-3 pt-2">
                        <button type="submit" disabled={isPending} className="bg-blue-800 text-white px-6 py-2 rounded-lg text-sm hover:bg-blue-900 disabled:opacity-60 transition-colors">{isPending ? 'Создание...' : 'Создать'}</button>
                        <button type="button" onClick={() => navigate('/work-orders')} className="px-6 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50">Отмена</button>
                    </div>
                </form>
            </div>
        </div>
    );
}
