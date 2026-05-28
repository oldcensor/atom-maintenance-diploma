import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth, hasAtLeastRole } from '../auth/AuthContext';
import Loading from '../utils/Loading';
import { useGetEquipmentQuery, useGetEquipmentTypesQuery, useGetDepartmentsQuery } from '../store/apiSlice';

const STATUS_LABELS = { active: 'Активно', inactive: 'Неактивно', under_maintenance: 'На ремонте', decommissioned: 'Списано' };
const STATUS_COLORS = {
    active: 'bg-green-100 text-green-700',
    inactive: 'bg-gray-100 text-gray-600',
    under_maintenance: 'bg-yellow-100 text-yellow-700',
    decommissioned: 'bg-red-100 text-red-600',
};

export default function EquipmentList() {
    const navigate = useNavigate();
    const { user } = useAuth();
    const { data: equipment = [], isLoading: isPending, error } = useGetEquipmentQuery();
    const { data: types = [] } = useGetEquipmentTypesQuery();
    const { data: departments = [] } = useGetDepartmentsQuery();

    const [filterStatus, setFilterStatus] = useState('');
    const [filterType, setFilterType] = useState('');
    const [filterDept, setFilterDept] = useState('');
    const [filterMine, setFilterMine] = useState(false);

    const typeMap = Object.fromEntries(types.map(t => [t.id, t.name]));
    const deptMap = Object.fromEntries(departments.map(d => [d.id, d.name]));

    const filtered = equipment.filter(e => {
        if (filterMine && e.responsible_id !== user?.id) return false;
        if (filterStatus && e.status !== filterStatus) return false;
        if (filterType && String(e.equipment_type_id) !== filterType) return false;
        if (filterDept && String(e.department_id) !== filterDept) return false;
        return true;
    });

    const hasFilters = filterStatus || filterType || filterDept || filterMine;

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-800">Оборудование</h1>
                {hasAtLeastRole(user?.role, 'manager') && (
                    <button onClick={() => navigate('/equipment/create')}
                        className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900 transition-colors">
                        + Создать
                    </button>
                )}
            </div>
            <div className="flex flex-wrap gap-3">
                <button
                    onClick={() => setFilterMine(v => !v)}
                    className={`px-3 py-2 rounded-lg text-sm border transition-colors ${filterMine ? 'bg-blue-800 text-white border-blue-800' : 'border-gray-300 text-gray-600 hover:border-gray-400'}`}>
                    Моё оборудование
                </button>
                <select value={filterStatus} onChange={e => setFilterStatus(e.target.value)}
                    className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Все статусы</option>
                    {Object.entries(STATUS_LABELS).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
                </select>
                <select value={filterType} onChange={e => setFilterType(e.target.value)}
                    className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Все типы</option>
                    {types.map(t => <option key={t.id} value={String(t.id)}>{t.name}</option>)}
                </select>
                <select value={filterDept} onChange={e => setFilterDept(e.target.value)}
                    className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">Все подразделения</option>
                    {departments.map(d => <option key={d.id} value={String(d.id)}>{d.name}</option>)}
                </select>
                {hasFilters && (
                    <button onClick={() => { setFilterStatus(''); setFilterType(''); setFilterDept(''); setFilterMine(false); }}
                        className="text-sm text-gray-500 hover:text-gray-700 px-2">Сбросить</button>
                )}
            </div>
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                {isPending ? (
                    <div className="flex justify-center py-12"><Loading /></div>
                ) : error ? (
                    <div className="px-5 py-8 text-center text-red-500">{error?.data?.error || 'Ошибка загрузки'}</div>
                ) : filtered.length === 0 ? (
                    <div className="px-5 py-12 text-center text-gray-400">Нет данных</div>
                ) : (
                    <div className="overflow-x-auto"><table className="w-full text-sm">
                        <thead className="bg-gray-50 text-gray-500 text-xs uppercase"><tr>
                            <th className="px-5 py-3 text-left">ID</th>
                            <th className="px-5 py-3 text-left">Название</th>
                            <th className="px-5 py-3 text-left">Серийный №</th>
                            <th className="px-5 py-3 text-left">Тип</th>
                            <th className="px-5 py-3 text-left">Подразделение</th>
                            <th className="px-5 py-3 text-left">Статус</th>
                        </tr></thead>
                        <tbody className="divide-y divide-gray-100">{filtered.map(e => (
                            <tr key={e.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/equipment/${e.id}`)}>
                                <td className="px-5 py-3 text-gray-400">#{e.id}</td>
                                <td className="px-5 py-3 font-medium text-gray-800">{e.name}</td>
                                <td className="px-5 py-3 font-mono text-gray-500">{e.serial_number}</td>
                                <td className="px-5 py-3 text-gray-600">{typeMap[e.equipment_type_id] || `#${e.equipment_type_id}`}</td>
                                <td className="px-5 py-3 text-gray-600">{deptMap[e.department_id] || '—'}</td>
                                <td className="px-5 py-3"><span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLORS[e.status] || 'bg-gray-100 text-gray-600'}`}>{STATUS_LABELS[e.status] || e.status}</span></td>
                            </tr>
                        ))}</tbody>
                    </table></div>
                )}
            </div>
        </div>
    );
}
