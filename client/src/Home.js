import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Loading from './utils/Loading';
import { useAuth } from './auth/AuthContext';
import {
    useGetWorkOrdersQuery,
    useGetEquipmentQuery,
    useGetDepartmentsQuery,
} from './store/apiSlice';

const STATUS_LABELS = { open: 'Открыт', in_progress: 'В работе', completed: 'Завершён', cancelled: 'Отменён' };
const WO_TYPE_LABELS = { emergency: 'Аварийный', corrective: 'Корректирующий', planned: 'Плановый' };
const WO_TYPE_COLORS = { emergency: 'bg-red-100 text-red-700', corrective: 'bg-yellow-100 text-yellow-700', planned: 'bg-blue-100 text-blue-700' };

function daysSince(dateStr) {
    return Math.floor((Date.now() - new Date(dateStr).getTime()) / 86400000);
}

const StatCard = ({ label, value, color }) => (
    <div className={`bg-white rounded-xl border-l-4 ${color} shadow-sm p-5 flex-1 min-w-[160px]`}>
        <p className="text-sm text-gray-500 mb-1">{label}</p>
        <p className="text-3xl font-bold text-gray-800">{value}</p>
    </div>
);

const Home = () => {
    const navigate = useNavigate();
    const { user } = useAuth();
    const [deptFilter, setDeptFilter] = useState('');

    const { data: workOrders = [], isLoading: woLoading } = useGetWorkOrdersQuery();
    const { data: equipment = [], isLoading: eqLoading } = useGetEquipmentQuery();
    const { data: departments = [] } = useGetDepartmentsQuery();

    const isLoading = woLoading || eqLoading;

    // Оборудование выбранного цеха (или всё)
    const filteredEquipment = deptFilter
        ? equipment.filter(e => String(e.department_id) === deptFilter)
        : equipment;
    const filteredEqIds = new Set(filteredEquipment.map(e => e.id));

    // Наряды и регламенты фильтруются через оборудование цеха
    const openOrders = workOrders.filter(w =>
        (w.status === 'open' || w.status === 'in_progress') &&
        (!deptFilter || filteredEqIds.has(w.equipment_id))
    );
    const overdueWorkOrders = workOrders.filter(w =>
        (!deptFilter || filteredEqIds.has(w.equipment_id)) &&
        ((w.status === 'open' && daysSince(w.created_at) > 3) ||
         (w.status === 'in_progress' && daysSince(w.created_at) > 7))
    );
    const underRepair = filteredEquipment.filter(e => e.status === 'under_maintenance');

    const eqMap = Object.fromEntries(equipment.map(e => [e.id, e.name]));

    // Мои активные наряды
    const myOrders = workOrders.filter(w =>
        w.assigned_to === user?.id &&
        (w.status === 'open' || w.status === 'in_progress')
    );

    if (isLoading) return <div className="flex justify-center items-center h-64"><Loading /></div>;

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between flex-wrap gap-3">
                <h1 className="text-2xl font-bold text-gray-800">Сводка</h1>
                <select
                    value={deptFilter}
                    onChange={e => setDeptFilter(e.target.value)}
                    className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 min-w-[200px]"
                >
                    <option value="">Все подразделения</option>
                    {departments.map(d => (
                        <option key={d.id} value={String(d.id)}>{d.name}</option>
                    ))}
                </select>
            </div>

            <div className="flex flex-wrap gap-4">
                <StatCard label="Открытые наряды" value={openOrders.length} color="border-blue-500" />
                <StatCard label="Просроченные наряды" value={overdueWorkOrders.length} color="border-red-500" />
                <StatCard label="Оборудование на ремонте" value={underRepair.length} color="border-yellow-500" />
            </div>

            {/* Мои задачи */}
            {myOrders.length > 0 && (
                <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden ring-1 ring-blue-200">
                    <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between bg-blue-50">
                        <h2 className="font-semibold text-blue-800">Мои задачи ({myOrders.length})</h2>
                        <button onClick={() => navigate('/work-orders')} className="text-sm text-blue-600 hover:text-blue-800">Все наряды →</button>
                    </div>
                    <div className="overflow-x-auto"><table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase"><tr>
                            <th className="px-5 py-3 text-left">ID</th>
                            <th className="px-5 py-3 text-left">Тип</th>
                            <th className="px-5 py-3 text-left">Заголовок</th>
                            <th className="px-5 py-3 text-left">Оборудование</th>
                            <th className="px-5 py-3 text-left">Статус</th>
                        </tr></thead>
                        <tbody className="divide-y divide-gray-100">{myOrders.map(wo => (
                            <tr key={wo.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/work-orders/${wo.id}`)}>
                                <td className="px-5 py-3 text-gray-500">#{wo.id}</td>
                                <td className="px-5 py-3">
                                    {wo.work_type
                                        ? <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${WO_TYPE_COLORS[wo.work_type] || 'bg-gray-100 text-gray-500'}`}>{WO_TYPE_LABELS[wo.work_type] || wo.work_type}</span>
                                        : <span className="text-gray-300 text-xs">—</span>
                                    }
                                </td>
                                <td className="px-5 py-3 font-medium text-gray-800">{wo.title}</td>
                                <td className="px-5 py-3 text-gray-600">{eqMap[wo.equipment_id] || `#${wo.equipment_id}`}</td>
                                <td className="px-5 py-3">
                                    <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${wo.status === 'open' ? 'bg-blue-100 text-blue-700' : 'bg-yellow-100 text-yellow-700'}`}>
                                        {STATUS_LABELS[wo.status] || wo.status}
                                    </span>
                                </td>
                            </tr>
                        ))}</tbody>
                    </table></div>
                </div>
            )}

            {/* Наряды */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
                    <h2 className="font-semibold text-gray-800">Наряды, требующие внимания</h2>
                    <button onClick={() => navigate('/work-orders')} className="text-sm text-blue-600 hover:text-blue-800">Все наряды →</button>
                </div>
                {openOrders.length === 0 ? (
                    <div className="px-5 py-8 text-center text-gray-400 text-sm">Нет открытых нарядов</div>
                ) : (
                    <div className="overflow-x-auto"><table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase"><tr>
                            <th className="px-5 py-3 text-left">ID</th>
                            <th className="px-5 py-3 text-left">Заголовок</th>
                            <th className="px-5 py-3 text-left">Оборудование</th>
                            <th className="px-5 py-3 text-left">Статус</th>
                            <th className="px-5 py-3 text-left">Создан</th>
                        </tr></thead>
                        <tbody className="divide-y divide-gray-100">{openOrders.slice(0, 10).map(wo => (
                            <tr key={wo.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/work-orders/${wo.id}`)}>
                                <td className="px-5 py-3 text-gray-500">#{wo.id}</td>
                                <td className="px-5 py-3 font-medium text-gray-800">{wo.title}</td>
                                <td className="px-5 py-3 text-gray-600">{eqMap[wo.equipment_id] || `#${wo.equipment_id}`}</td>
                                <td className="px-5 py-3">
                                    <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${wo.status === 'open' ? 'bg-blue-100 text-blue-700' : 'bg-yellow-100 text-yellow-700'}`}>
                                        {STATUS_LABELS[wo.status] || wo.status}
                                    </span>
                                </td>
                                <td className="px-5 py-3 text-gray-500">{new Date(wo.created_at).toLocaleDateString('ru-RU')}</td>
                            </tr>
                        ))}</tbody>
                    </table></div>
                )}
            </div>

            {/* Просроченные наряды */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
                    <h2 className="font-semibold text-gray-800">Просроченные наряды</h2>
                    <button onClick={() => navigate('/work-orders')} className="text-sm text-blue-600 hover:text-blue-800">Все наряды →</button>
                </div>
                {overdueWorkOrders.length === 0 ? (
                    <div className="px-5 py-8 text-center text-gray-400 text-sm">Нет просроченных нарядов</div>
                ) : (
                    <div className="overflow-x-auto"><table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase"><tr>
                            <th className="px-5 py-3 text-left">ID</th>
                            <th className="px-5 py-3 text-left">Тип</th>
                            <th className="px-5 py-3 text-left">Заголовок</th>
                            <th className="px-5 py-3 text-left">Оборудование</th>
                            <th className="px-5 py-3 text-left">Дней открыт</th>
                        </tr></thead>
                        <tbody className="divide-y divide-gray-100">{overdueWorkOrders.slice(0, 10).map(wo => (
                                <tr key={wo.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/work-orders/${wo.id}`)}>
                                    <td className="px-5 py-3 text-gray-500">#{wo.id}</td>
                                    <td className="px-5 py-3">
                                        {wo.work_type
                                            ? <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${WO_TYPE_COLORS[wo.work_type] || 'bg-gray-100 text-gray-500'}`}>{WO_TYPE_LABELS[wo.work_type] || wo.work_type}</span>
                                            : <span className="text-gray-300 text-xs">—</span>
                                        }
                                    </td>
                                    <td className="px-5 py-3 font-medium text-gray-800">{wo.title}</td>
                                    <td className="px-5 py-3 text-gray-600">{eqMap[wo.equipment_id] || `#${wo.equipment_id}`}</td>
                                    <td className="px-5 py-3"><span className="text-red-600 font-medium">⚠ {daysSince(wo.created_at)} дн.</span></td>
                                </tr>
                        ))}</tbody>
                    </table></div>
                )}
            </div>

            {/* Оборудование на ремонте */}
            {underRepair.length > 0 && (
                <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                    <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
                        <h2 className="font-semibold text-gray-800">Оборудование на ремонте</h2>
                        <button onClick={() => navigate('/equipment')} className="text-sm text-blue-600 hover:text-blue-800">Весь реестр →</button>
                    </div>
                    <div className="overflow-x-auto"><table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase"><tr>
                            <th className="px-5 py-3 text-left">ID</th>
                            <th className="px-5 py-3 text-left">Название</th>
                            <th className="px-5 py-3 text-left">Серийный №</th>
                        </tr></thead>
                        <tbody className="divide-y divide-gray-100">{underRepair.map(e => (
                            <tr key={e.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/equipment/${e.id}`)}>
                                <td className="px-5 py-3 text-gray-500">#{e.id}</td>
                                <td className="px-5 py-3 font-medium text-gray-800">{e.name}</td>
                                <td className="px-5 py-3 text-gray-500 font-mono">{e.serial_number}</td>
                            </tr>
                        ))}</tbody>
                    </table></div>
                </div>
            )}
        </div>
    );
};

export default Home;
