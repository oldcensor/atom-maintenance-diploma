import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import {
    useGetEquipmentByIdQuery,
    useGetEquipmentTypesQuery,
    useGetDepartmentsQuery,
    useGetEmployeesQuery,
    useGetWorkOrdersQuery,
    useGetMaintenanceSchedulesQuery,
    useUpdateEquipmentMutation,
} from '../store/apiSlice';
import { useAuth, hasAtLeastRole } from '../auth/AuthContext';
import Loading from '../utils/Loading';
import { toast } from 'react-toastify';
import Config from '../utils/Config';

const ROLE_LABELS = { technician: 'Техник', engineer: 'Инженер', manager: 'Менеджер', admin: 'Администратор' };
const ASSIGNABLE_ROLES = ['technician', 'engineer'];

const STATUS_LABELS = {
    active: 'Активно',
    inactive: 'Неактивно',
    under_maintenance: 'На ремонте',
    decommissioned: 'Списано',
};

const STATUS_COLORS = {
    active: 'bg-green-100 text-green-700',
    inactive: 'bg-gray-100 text-gray-600',
    under_maintenance: 'bg-yellow-100 text-yellow-700',
    decommissioned: 'bg-red-100 text-red-600',
};

const WO_STATUS_LABELS = {
    open: 'Открыт',
    in_progress: 'В работе',
    completed: 'Завершён',
    cancelled: 'Отменён',
};

const SCH_STATUS_LABELS = {
    scheduled: 'Запланирован',
    in_progress: 'В работе',
    completed: 'Завершён',
    cancelled: 'Отменён',
};

const INTERVAL_LABELS = { days: 'дней', operating_hours: 'моточасов', cycles: 'циклов' };

export default function EquipmentDetails() {
    const { id } = useParams();
    const navigate = useNavigate();
    const { user } = useAuth();

    const { data: eq, isLoading, error } = useGetEquipmentByIdQuery(id);
    const { data: types } = useGetEquipmentTypesQuery();
    const { data: departments } = useGetDepartmentsQuery();
    const { data: workOrders } = useGetWorkOrdersQuery();
    const { data: schedules } = useGetMaintenanceSchedulesQuery();
    const [updateEq, { isLoading: updating }] = useUpdateEquipmentMutation();

    const [editing, setEditing] = useState(false);
    const [form, setForm] = useState(null);

    const [telemetry, setTelemetry] = useState(null);
    const [history, setHistory] = useState([]);
    const intervalRef = useRef(null);

    useEffect(() => {
        const SIMULATOR = Config.endpoints.simulatorUrl;
        const fetchTelemetry = () => {
            fetch(`${SIMULATOR}/api/v1/telemetry/${id}`)
                .then(r => r.ok ? r.json() : null)
                .then(data => setTelemetry(data))
                .catch(() => setTelemetry(null));
            fetch(`${SIMULATOR}/api/v1/telemetry/${id}/history?n=60&interval=300`)
                .then(r => r.ok ? r.json() : null)
                .then(pts => {
                    if (!pts) return;
                    setHistory(pts.map(p => ({
                        t: new Date(p.timestamp).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit' }),
                        v: p.value,
                        unit: p.unit,
                    })));
                })
                .catch(() => {});
        };
        fetchTelemetry();
        intervalRef.current = setInterval(fetchTelemetry, 5000);
        return () => clearInterval(intervalRef.current);
    }, [id]);

    const { data: employees } = useGetEmployeesQuery();

    const typeMap = Object.fromEntries((types || []).map(t => [t.id, t.name]));
    const deptMap = Object.fromEntries((departments || []).map(d => [d.id, d.name]));
    const empMap = Object.fromEntries((employees || []).map(e => [e.id, e.full_name]));

    const eqWorkOrders = (workOrders || []).filter(w => w.equipment_id === parseInt(id, 10));
    const eqSchedules = (schedules || []).filter(s => s.equipment_id === parseInt(id, 10));

    const startEdit = () => {
        setForm({
            name: eq.name,
            description: eq.description || '',
            serial_number: eq.serial_number,
            equipment_type_id: String(eq.equipment_type_id),
            department_id: eq.department_id ? String(eq.department_id) : '',
            responsible_id: eq.responsible_id ? String(eq.responsible_id) : '',
            status: eq.status,
        });
        setEditing(true);
    };

    const handleSave = async (e) => {
        e.preventDefault();
        try {
            await updateEq({
                id,
                name: form.name,
                description: form.description,
                serial_number: form.serial_number,
                equipment_type_id: parseInt(form.equipment_type_id, 10),
                department_id: form.department_id ? parseInt(form.department_id, 10) : null,
                responsible_id: form.responsible_id ? parseInt(form.responsible_id, 10) : null,
                status: form.status,
            }).unwrap();
            toast.success('Сохранено');
            setEditing(false);
        } catch (err) {
            toast.error(err.data?.error || err.message || 'Ошибка');
        }
    };

    const handleDecommission = async () => {
        if (!window.confirm('Перевести оборудование в статус "Списано"? Все активные регламенты будут деактивированы.')) return;
        try {
            await updateEq({
                id,
                name: eq.name,
                description: eq.description || '',
                serial_number: eq.serial_number,
                equipment_type_id: eq.equipment_type_id,
                department_id: eq.department_id || null,
                responsible_id: eq.responsible_id || null,
                status: 'decommissioned',
            }).unwrap();
            toast.success('Оборудование списано');
        } catch (err) {
            toast.error(err.data?.error || err.message || 'Ошибка');
        }
    };

    const set = (field) => (e) => setForm(f => ({ ...f, [field]: e.target.value }));

    if (isLoading) return <div className="flex justify-center py-12"><Loading /></div>;
    if (error) return <div className="text-red-500 p-4">{error.data?.error || 'Ошибка загрузки'}</div>;
    if (!eq) return null;

    return (
        <div className="space-y-6 max-w-4xl">
            {/* Заголовок */}
            <div className="flex items-center gap-3">
                <button onClick={() => navigate('/equipment')} className="text-gray-400 hover:text-gray-600">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
                    </svg>
                </button>
                <h1 className="text-2xl font-bold text-gray-800">{eq.name}</h1>
                <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLORS[eq.status] || 'bg-gray-100 text-gray-600'}`}>
                    {STATUS_LABELS[eq.status] || eq.status}
                </span>
            </div>

            {/* Карточка */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                {!editing ? (
                    <>
                        <div className="grid grid-cols-2 gap-4 text-sm mb-6">
                            <div><span className="text-gray-500">Серийный №:</span> <span className="font-mono font-medium ml-2">{eq.serial_number}</span></div>
                            <div><span className="text-gray-500">Тип:</span> <span className="ml-2">{typeMap[eq.equipment_type_id] || `#${eq.equipment_type_id}`}</span></div>
                            <div><span className="text-gray-500">Подразделение:</span> <span className="ml-2">{eq.department_id ? deptMap[eq.department_id] || `#${eq.department_id}` : '—'}</span></div>
                            <div><span className="text-gray-500">Ответственный:</span> <span className="ml-2">{eq.responsible_id ? empMap[eq.responsible_id] || `#${eq.responsible_id}` : '—'}</span></div>
                            <div><span className="text-gray-500">Создано:</span> <span className="ml-2">{new Date(eq.created_at).toLocaleDateString('ru-RU')}</span></div>
                            {eq.description && (
                                <div className="col-span-2"><span className="text-gray-500">Описание:</span> <span className="ml-2 whitespace-pre-wrap">{eq.description}</span></div>
                            )}
                        </div>

                        {hasAtLeastRole(user?.role, 'manager') && (
                            <div className="flex gap-3">
                                <button
                                    onClick={startEdit}
                                    className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900 transition-colors"
                                >
                                    Редактировать
                                </button>
                                {eq.status !== 'decommissioned' && (
                                    <button
                                        onClick={handleDecommission}
                                        className="px-4 py-2 border border-red-300 text-red-600 rounded-lg text-sm hover:bg-red-50 transition-colors"
                                    >
                                        Вывести из эксплуатации
                                    </button>
                                )}
                            </div>
                        )}
                    </>
                ) : (
                    <form onSubmit={handleSave} className="space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="col-span-2">
                                <label className="block text-sm font-medium text-gray-700 mb-1">Название *</label>
                                <input required maxLength={100} value={form.name} onChange={set('name')}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Серийный номер *</label>
                                <input required maxLength={50} value={form.serial_number} onChange={set('serial_number')}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500" />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Статус</label>
                                <select value={form.status} onChange={set('status')}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                    {Object.entries(STATUS_LABELS).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Тип оборудования *</label>
                                <select required value={form.equipment_type_id} onChange={set('equipment_type_id')}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                    <option value="">Выберите тип</option>
                                    {(types || []).map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Подразделение</label>
                                <select value={form.department_id} onChange={set('department_id')}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                    <option value="">Не указано</option>
                                    {(departments || []).map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Ответственный</label>
                                <select value={form.responsible_id} onChange={set('responsible_id')}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                    <option value="">Не назначен</option>
                                    {(employees || []).filter(e => ASSIGNABLE_ROLES.includes(e.role)).map(e => <option key={e.id} value={e.id}>{e.full_name} ({ROLE_LABELS[e.role] ?? e.role})</option>)}
                                </select>
                            </div>
                            <div className="col-span-2">
                                <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
                                <textarea rows={3} maxLength={1000} value={form.description} onChange={set('description')}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none" />
                            </div>
                        </div>
                        <div className="flex gap-3">
                            <button type="submit" disabled={updating}
                                className="bg-blue-800 text-white px-6 py-2 rounded-lg text-sm hover:bg-blue-900 disabled:opacity-60 transition-colors">
                                {updating ? 'Сохранение...' : 'Сохранить'}
                            </button>
                            <button type="button" onClick={() => setEditing(false)}
                                className="px-6 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50">
                                Отмена
                            </button>
                        </div>
                    </form>
                )}
            </div>

            {/* Телеметрия симулятора */}
            {telemetry && (
                <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
                    <div className="flex items-center justify-between mb-3">
                        <h2 className="font-semibold text-gray-800">Наработка (симулятор)</h2>
                        <span className="inline-flex items-center gap-1.5 text-xs text-green-600">
                            <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                            live
                        </span>
                    </div>
                    <div className="flex items-end gap-2 mb-4">
                        <span className="text-3xl font-bold text-gray-900 tabular-nums">
                            {telemetry.current_value.toLocaleString('ru-RU')}
                        </span>
                        <span className="text-gray-500 mb-1">{telemetry.unit}</span>
                        <span className="text-xs text-gray-400 mb-1 ml-2">
                            Обновлено: {new Date(telemetry.timestamp).toLocaleTimeString('ru-RU')}
                        </span>
                    </div>
                    {history.length > 0 && (
                        <ResponsiveContainer width="100%" height={180}>
                            <AreaChart data={history} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
                                <defs>
                                    <linearGradient id="telGrad" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.25} />
                                        <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                                <XAxis
                                    dataKey="t"
                                    tick={{ fontSize: 10, fill: '#9ca3af' }}
                                    interval={Math.floor(history.length / 5)}
                                    tickLine={false}
                                    axisLine={false}
                                />
                                <YAxis
                                    tick={{ fontSize: 10, fill: '#9ca3af' }}
                                    tickLine={false}
                                    axisLine={false}
                                    width={55}
                                    tickFormatter={v => v.toLocaleString('ru-RU')}
                                />
                                <Tooltip
                                    formatter={(v) => [v.toLocaleString('ru-RU') + ' ' + (history[0]?.unit ?? ''), 'Наработка']}
                                    labelFormatter={(l) => l}
                                    contentStyle={{ fontSize: 12, borderRadius: 8, border: '1px solid #e5e7eb' }}
                                />
                                <Area
                                    type="monotone"
                                    dataKey="v"
                                    stroke="#3b82f6"
                                    strokeWidth={2}
                                    fill="url(#telGrad)"
                                    dot={false}
                                    activeDot={{ r: 4 }}
                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    )}
                </div>
            )}

            {/* Регламенты */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
                    <h2 className="font-semibold text-gray-800">Регламенты</h2>
                    {hasAtLeastRole(user?.role, 'engineer') && eq.status !== 'decommissioned' && (
                        <button
                            onClick={() => navigate(`/schedules/create?equipment_id=${id}`)}
                            className="text-sm text-blue-600 hover:text-blue-800"
                        >
                            + Добавить
                        </button>
                    )}
                </div>
                {eqSchedules.length === 0 ? (
                    <div className="px-5 py-6 text-center text-gray-400 text-sm">Нет регламентов</div>
                ) : (
                    <table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase">
                            <tr>
                                <th className="px-5 py-3 text-left">ID</th>
                                <th className="px-5 py-3 text-left">Статус</th>
                                <th className="px-5 py-3 text-left">Интервал</th>
                                <th className="px-5 py-3 text-left">Следующее ТО</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-100">
                            {eqSchedules.map(s => (
                                <tr key={s.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/schedules/${s.id}`)}>
                                    <td className="px-5 py-3 text-gray-400">#{s.id}</td>
                                    <td className="px-5 py-3 text-gray-600">{SCH_STATUS_LABELS[s.status] || s.status}</td>
                                    <td className="px-5 py-3 text-gray-600">
                                        {s.interval_value ? `${s.interval_value} ${INTERVAL_LABELS[s.interval_unit] || s.interval_unit}` : '—'}
                                    </td>
                                    <td className="px-5 py-3">
                                        {s.next_due_at ? (
                                            <span className={new Date(s.next_due_at) < new Date() ? 'text-red-600 font-medium' : 'text-gray-600'}>
                                                {new Date(s.next_due_at).toLocaleDateString('ru-RU')}
                                            </span>
                                        ) : '—'}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>

            {/* Наряды */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
                    <h2 className="font-semibold text-gray-800">Наряды-задания</h2>
                    {hasAtLeastRole(user?.role, 'engineer') && eq.status !== 'decommissioned' && (
                        <button
                            onClick={() => navigate(`/work-orders/create?equipment_id=${id}`)}
                            className="text-sm text-blue-600 hover:text-blue-800"
                        >
                            + Создать
                        </button>
                    )}
                </div>
                {eqWorkOrders.length === 0 ? (
                    <div className="px-5 py-6 text-center text-gray-400 text-sm">Нет нарядов</div>
                ) : (
                    <table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase">
                            <tr>
                                <th className="px-5 py-3 text-left">ID</th>
                                <th className="px-5 py-3 text-left">Заголовок</th>
                                <th className="px-5 py-3 text-left">Статус</th>
                                <th className="px-5 py-3 text-left">Создан</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-100">
                            {eqWorkOrders.map(w => (
                                <tr key={w.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/work-orders/${w.id}`)}>
                                    <td className="px-5 py-3 text-gray-400">#{w.id}</td>
                                    <td className="px-5 py-3 font-medium text-gray-800">{w.title}</td>
                                    <td className="px-5 py-3 text-gray-600">{WO_STATUS_LABELS[w.status] || w.status}</td>
                                    <td className="px-5 py-3 text-gray-500">{new Date(w.created_at).toLocaleDateString('ru-RU')}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>
        </div>
    );
}
