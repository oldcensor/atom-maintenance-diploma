import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth, hasAtLeastRole } from '../auth/AuthContext';
import Loading from '../utils/Loading';
import { useGetWorkOrdersQuery, useGetEquipmentQuery, useGetEmployeesQuery } from '../store/apiSlice';


const STATUS_LABELS = { open: 'Открыт', in_progress: 'В работе', completed: 'Завершён', cancelled: 'Отменён' };
const STATUS_COLORS = { open: 'bg-blue-100 text-blue-700', in_progress: 'bg-yellow-100 text-yellow-700', completed: 'bg-green-100 text-green-700', cancelled: 'bg-gray-100 text-gray-500' };

const WO_TYPE_LABELS = { emergency: 'Аварийный', corrective: 'Корректирующий', planned: 'Плановый' };
const WO_TYPE_COLORS = {
    emergency:  'bg-red-100 text-red-700',
    corrective: 'bg-yellow-100 text-yellow-700',
    planned:    'bg-blue-100 text-blue-700',
};

export default function WorkOrderList() {
    const navigate = useNavigate();
    const { user } = useAuth();
    const { data: workOrders = [], isLoading: isPending, error } = useGetWorkOrdersQuery();
    const { data: equipment = [] } = useGetEquipmentQuery();
    const { data: employees = [] } = useGetEmployeesQuery();

    const [filterStatus, setFilterStatus] = useState('');
    const [filterEq, setFilterEq] = useState('');
    const [filterAssignee, setFilterAssignee] = useState('');
    const [filterType, setFilterType] = useState('');
    const [filterMine, setFilterMine] = useState(false);

    const eqMap = Object.fromEntries(equipment.map(e => [e.id, e.name]));
    const empMap = Object.fromEntries(employees.map(e => [e.id, e.full_name]));

    const filtered = workOrders.filter(w => {
        if (filterMine && w.assigned_to !== user?.id) return false;
        if (filterStatus && w.status !== filterStatus) return false;
        if (filterEq && String(w.equipment_id) !== filterEq) return false;
        if (filterAssignee && String(w.assigned_to) !== filterAssignee) return false;
        if (filterType && w.work_type !== filterType) return false;
        return true;
    });

    const hasFilters = filterStatus || filterEq || filterAssignee || filterType || filterMine;

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-800">Наряды-задания</h1>
                {hasAtLeastRole(user?.role, 'engineer') && (
                    <button onClick={() => navigate('/work-orders/create')} className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900 transition-colors">+ Создать</button>
                )}
            </div>
            <div className="flex flex-wrap gap-3">
                <button
                    onClick={() => setFilterMine(v => !v)}
                    className={`px-3 py-2 rounded-lg text-sm border transition-colors ${filterMine ? 'bg-blue-800 text-white border-blue-800' : 'border-gray-300 text-gray-600 hover:border-gray-400'}`}>
                    Мои наряды
                </button>
                <select value={filterType} onChange={e => setFilterType(e.target.value)} className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Все типы</option>
                    {Object.entries(WO_TYPE_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                </select>
                <select value={filterStatus} onChange={e => setFilterStatus(e.target.value)} className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Все статусы</option>
                    {Object.entries(STATUS_LABELS).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
                </select>
                <select value={filterEq} onChange={e => setFilterEq(e.target.value)} className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Всё оборудование</option>
                    {equipment.map(e => <option key={e.id} value={String(e.id)}>{e.name}</option>)}
                </select>
                <select value={filterAssignee} onChange={e => setFilterAssignee(e.target.value)} className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Все исполнители</option>
                    {employees.map(e => <option key={e.id} value={String(e.id)}>{e.full_name}</option>)}
                </select>
                {hasFilters && (
                    <button onClick={() => { setFilterStatus(''); setFilterEq(''); setFilterAssignee(''); setFilterType(''); setFilterMine(false); }} className="text-sm text-gray-500 hover:text-gray-700 px-2">Сбросить</button>
                )}
            </div>
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                {isPending ? <div className="flex justify-center py-12"><Loading /></div>
                : error ? <div className="px-5 py-8 text-center text-red-500">{error?.data?.error || 'Ошибка загрузки'}</div>
                : filtered.length === 0 ? <div className="px-5 py-12 text-center text-gray-400">Нет данных</div>
                : (
                    <div className="overflow-x-auto"><table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase"><tr>
                            <th className="px-5 py-3 text-left">ID</th>
                            <th className="px-5 py-3 text-left">Тип</th>
                            <th className="px-5 py-3 text-left">Заголовок</th>
                            <th className="px-5 py-3 text-left">Оборудование</th>
                            <th className="px-5 py-3 text-left">Исполнитель</th>
                            <th className="px-5 py-3 text-left">Статус</th>
                            <th className="px-5 py-3 text-left">Создан</th>
                        </tr></thead>
                        <tbody className="divide-y divide-gray-100">{filtered.map(w => (
                                <tr key={w.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/work-orders/${w.id}`)}>
                                    <td className="px-5 py-3 text-gray-400">#{w.id}</td>
                                    <td className="px-5 py-3">
                                        {w.work_type
                                            ? <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${WO_TYPE_COLORS[w.work_type] || 'bg-gray-100 text-gray-500'}`}>{WO_TYPE_LABELS[w.work_type] || w.work_type}</span>
                                            : <span className="text-gray-300 text-xs">—</span>
                                        }
                                    </td>
                                    <td className="px-5 py-3 font-medium text-gray-800">{w.title}</td>
                                    <td className="px-5 py-3 text-gray-600">{eqMap[w.equipment_id] || `#${w.equipment_id}`}</td>
                                    <td className="px-5 py-3 text-gray-600">{w.assigned_to ? empMap[w.assigned_to] || `#${w.assigned_to}` : '—'}</td>
                                    <td className="px-5 py-3"><span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLORS[w.status] || 'bg-gray-100'}`}>{STATUS_LABELS[w.status] || w.status}</span></td>
                                    <td className="px-5 py-3 text-gray-500">{new Date(w.created_at).toLocaleDateString('ru-RU')}</td>
                                </tr>
                        ))}</tbody>
                    </table></div>
                )}
            </div>
        </div>
    );
}
