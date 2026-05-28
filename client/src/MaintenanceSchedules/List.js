import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth, hasAtLeastRole } from '../auth/AuthContext';
import Loading from '../utils/Loading';
import { useGetMaintenanceSchedulesQuery, useGetEquipmentQuery } from '../store/apiSlice';

const STATUS_LABELS = { scheduled: 'Запланирован', in_progress: 'В работе', completed: 'Завершён', cancelled: 'Отменён' };
const STATUS_COLORS = { scheduled: 'bg-blue-100 text-blue-700', in_progress: 'bg-yellow-100 text-yellow-700', completed: 'bg-green-100 text-green-700', cancelled: 'bg-gray-100 text-gray-500' };
const INTERVAL_LABELS = { days: 'дн.', operating_hours: 'ч.', cycles: 'цикл.' };

export default function ScheduleList() {
    const navigate = useNavigate();
    const { user } = useAuth();
    const { data: schedules = [], isLoading: isPending, error } = useGetMaintenanceSchedulesQuery();
    const { data: equipment = [] } = useGetEquipmentQuery();

    const [filterStatus, setFilterStatus] = useState('');
    const [filterEq, setFilterEq] = useState('');
    const [filterMine, setFilterMine] = useState(false);

    const eqMap = Object.fromEntries(equipment.map(e => [e.id, e.name]));
    const myEqIds = new Set(equipment.filter(e => e.responsible_id === user?.id).map(e => e.id));
    const now = new Date();
    const filtered = schedules.filter(s => {
        if (filterMine && !myEqIds.has(s.equipment_id)) return false;
        if (filterStatus && s.status !== filterStatus) return false;
        if (filterEq && String(s.equipment_id) !== filterEq) return false;
        return true;
    });

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-800">Регламенты</h1>
                {hasAtLeastRole(user?.role, 'engineer') && (
                    <button onClick={() => navigate('/schedules/create')} className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900 transition-colors">+ Создать</button>
                )}
            </div>
            <div className="flex flex-wrap gap-3">
                <button
                    onClick={() => setFilterMine(v => !v)}
                    className={`px-3 py-2 rounded-lg text-sm border transition-colors ${filterMine ? 'bg-blue-800 text-white border-blue-800' : 'border-gray-300 text-gray-600 hover:border-gray-400'}`}>
                    Моё оборудование
                </button>
                <select value={filterStatus} onChange={e => setFilterStatus(e.target.value)} className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Все статусы</option>
                    {Object.entries(STATUS_LABELS).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
                </select>
                <select value={filterEq} onChange={e => setFilterEq(e.target.value)} className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Всё оборудование</option>
                    {equipment.map(e => <option key={e.id} value={String(e.id)}>{e.name}</option>)}
                </select>
                {(filterStatus || filterEq || filterMine) && (
                    <button onClick={() => { setFilterStatus(''); setFilterEq(''); setFilterMine(false); }} className="text-sm text-gray-500 hover:text-gray-700 px-2">Сбросить</button>
                )}
            </div>
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                {isPending ? <div className="flex justify-center py-12"><Loading /></div>
                : error ? <div className="px-5 py-8 text-center text-red-500">{error?.data?.error || 'Ошибка загрузки'}</div>
                : filtered.length === 0 ? <div className="px-5 py-12 text-center text-gray-400">Нет данных</div>
                : (
                    <div className="overflow-x-auto"><table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase"><tr>
                            <th className="px-5 py-3 text-left">ID</th><th className="px-5 py-3 text-left">Оборудование</th>
                            <th className="px-5 py-3 text-left">Статус</th><th className="px-5 py-3 text-left">Интервал</th>
                            <th className="px-5 py-3 text-left">Следующее ТО</th><th className="px-5 py-3 text-left">Запланировано</th>
                        </tr></thead>
                        <tbody className="divide-y divide-gray-100">{filtered.map(s => {
                            const overdue = s.next_due_at && new Date(s.next_due_at) < now && s.status !== 'completed' && s.status !== 'cancelled';
                            return (
                                <tr key={s.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/schedules/${s.id}`)}>
                                    <td className="px-5 py-3 text-gray-400">#{s.id}</td>
                                    <td className="px-5 py-3 font-medium text-gray-800">{eqMap[s.equipment_id] || `#${s.equipment_id}`}</td>
                                    <td className="px-5 py-3"><span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLORS[s.status] || 'bg-gray-100 text-gray-500'}`}>{STATUS_LABELS[s.status] || s.status}</span></td>
                                    <td className="px-5 py-3 text-gray-600">{s.interval_value ? `${s.interval_value} ${INTERVAL_LABELS[s.interval_unit] || s.interval_unit}` : '—'}</td>
                                    <td className="px-5 py-3">{s.next_due_at ? <span className={overdue ? 'text-red-600 font-medium' : 'text-gray-600'}>{overdue && '⚠ '}{new Date(s.next_due_at).toLocaleDateString('ru-RU')}</span> : '—'}</td>
                                    <td className="px-5 py-3 text-gray-500">{new Date(s.scheduled_at).toLocaleDateString('ru-RU')}</td>
                                </tr>
                            );
                        })}</tbody>
                    </table></div>
                )}
            </div>
        </div>
    );
}
