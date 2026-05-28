import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    useGetMaintenanceScheduleByIdQuery,
    useGetEquipmentQuery,
    useUpdateMaintenanceScheduleMutation,
    useDeleteMaintenanceScheduleMutation,
} from '../store/apiSlice';
import { useAuth, hasAtLeastRole } from '../auth/AuthContext';
import Loading from '../utils/Loading';
import { toast } from 'react-toastify';

const STATUS_LABELS = { scheduled: 'Запланирован', in_progress: 'В работе', completed: 'Завершён', cancelled: 'Отменён' };
const INTERVAL_LABELS = { days: 'дней', operating_hours: 'моточасов', cycles: 'циклов' };

export default function ScheduleDetails() {
    const { id } = useParams();
    const navigate = useNavigate();
    const { user } = useAuth();

    const { data: schedule, isLoading, error } = useGetMaintenanceScheduleByIdQuery(id);
    const { data: equipment } = useGetEquipmentQuery();
    const [updateSchedule, { isLoading: updating }] = useUpdateMaintenanceScheduleMutation();
    const [deleteSchedule] = useDeleteMaintenanceScheduleMutation();

    const [editing, setEditing] = useState(false);
    const [editForm, setEditForm] = useState({});

    const eqMap = Object.fromEntries((equipment || []).map(e => [e.id, e.name]));
    const canManage = hasAtLeastRole(user?.role, 'engineer');

    const startEdit = () => {
        setEditForm({
            description: schedule.description || '',
            scheduled_at: schedule.scheduled_at ? schedule.scheduled_at.slice(0, 16) : '',
            interval_unit: schedule.interval_unit || '',
            interval_value: schedule.interval_value || '',
        });
        setEditing(true);
    };

    const handleSave = async (e) => {
        e.preventDefault();
        try {
            const body = {
                id,
                equipment_id: schedule.equipment_id,
                scheduled_at: new Date(editForm.scheduled_at).toISOString(),
                description: editForm.description,
                assigned_to: null,
                status: schedule.status,
            };
            if (editForm.interval_unit) {
                body.interval_unit = editForm.interval_unit;
                body.interval_value = parseInt(editForm.interval_value, 10);
            } else {
                body.interval_unit = null;
                body.interval_value = null;
            }
            await updateSchedule(body).unwrap();
            toast.success('Регламент обновлён');
            setEditing(false);
        } catch (err) {
            toast.error(err.data?.error || err.message || 'Ошибка');
        }
    };

    const handleDeactivate = async () => {
        if (!window.confirm('Деактивировать регламент?')) return;
        try {
            await updateSchedule({
                id,
                equipment_id: schedule.equipment_id,
                scheduled_at: schedule.scheduled_at,
                description: schedule.description || '',
                assigned_to: null,
                status: 'cancelled',
                interval_unit: schedule.interval_unit || null,
                interval_value: schedule.interval_value || null,
            }).unwrap();
            toast.success('Регламент деактивирован');
        } catch (err) {
            toast.error(err.data?.error || err.message || 'Ошибка');
        }
    };

    const handleDelete = async () => {
        if (!window.confirm('Удалить регламент? Это действие необратимо.')) return;
        try {
            await deleteSchedule(id).unwrap();
            toast.success('Регламент удалён');
            navigate('/schedules');
        } catch (err) {
            toast.error(err.data?.error || err.message || 'Ошибка');
        }
    };

    if (isLoading) return <div className="flex justify-center py-12"><Loading /></div>;
    if (error) return <div className="text-red-500 p-4">{error.data?.error || 'Ошибка загрузки'}</div>;
    if (!schedule) return null;

    const overdue = schedule.next_due_at && new Date(schedule.next_due_at) < new Date()
        && schedule.status !== 'completed' && schedule.status !== 'cancelled';
    const canEdit = canManage && schedule.status !== 'cancelled' && schedule.status !== 'completed';

    return (
        <div className="space-y-6 max-w-3xl">
            <div className="flex items-center gap-3">
                <button onClick={() => navigate('/schedules')} className="text-gray-400 hover:text-gray-600">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
                    </svg>
                </button>
                <h1 className="text-2xl font-bold text-gray-800">Регламент #{schedule.id}</h1>
                <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${
                    schedule.status === 'completed' ? 'bg-green-100 text-green-700' :
                    schedule.status === 'cancelled' ? 'bg-gray-100 text-gray-500' :
                    schedule.status === 'in_progress' ? 'bg-yellow-100 text-yellow-700' :
                    'bg-blue-100 text-blue-700'
                }`}>
                    {STATUS_LABELS[schedule.status] || schedule.status}
                </span>
                {canEdit && !editing && (
                    <button onClick={startEdit} className="ml-auto text-sm text-blue-600 hover:text-blue-800">
                        Редактировать
                    </button>
                )}
            </div>

            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 space-y-4">
                {editing ? (
                    <form onSubmit={handleSave} className="space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="col-span-2">
                                <label className="block text-sm font-medium text-gray-700 mb-1">Дата планирования *</label>
                                <input type="datetime-local" required value={editForm.scheduled_at}
                                    onChange={e => setEditForm(f => ({ ...f, scheduled_at: e.target.value }))}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Единица интервала</label>
                                <div className="flex gap-3 mt-1">
                                    {[
                                        { value: '', label: 'Нет' },
                                        { value: 'days', label: 'Дни' },
                                        { value: 'operating_hours', label: 'Моточасы' },
                                        { value: 'cycles', label: 'Циклы' },
                                    ].map(opt => (
                                        <label key={opt.value} className="flex items-center gap-1 text-sm cursor-pointer">
                                            <input type="radio" name="edit_interval_unit" value={opt.value}
                                                checked={editForm.interval_unit === opt.value}
                                                onChange={e => setEditForm(f => ({ ...f, interval_unit: e.target.value, interval_value: e.target.value ? f.interval_value : '' }))}
                                                className="accent-blue-700" />
                                            {opt.label}
                                        </label>
                                    ))}
                                </div>
                            </div>
                            {editForm.interval_unit && (
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-1">Значение интервала *</label>
                                    <input type="number" required min={1} value={editForm.interval_value}
                                        onChange={e => setEditForm(f => ({ ...f, interval_value: e.target.value }))}
                                        className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                                </div>
                            )}
                            <div className="col-span-2">
                                <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
                                <textarea rows={3} maxLength={1000} value={editForm.description}
                                    onChange={e => setEditForm(f => ({ ...f, description: e.target.value }))}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none" />
                            </div>
                        </div>
                        <div className="flex gap-3 pt-2">
                            <button type="submit" disabled={updating}
                                className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900 disabled:opacity-60">
                                Сохранить
                            </button>
                            <button type="button" onClick={() => setEditing(false)}
                                className="px-4 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50">
                                Отмена
                            </button>
                        </div>
                    </form>
                ) : (
                    <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                            <span className="text-gray-500">Оборудование:</span>
                            <button className="ml-2 text-blue-600 hover:underline" onClick={() => navigate(`/equipment/${schedule.equipment_id}`)}>
                                {eqMap[schedule.equipment_id] || `#${schedule.equipment_id}`}
                            </button>
                        </div>
                        <div>
                            <span className="text-gray-500">Запланировано:</span>
                            <span className="ml-2">{new Date(schedule.scheduled_at).toLocaleDateString('ru-RU')}</span>
                        </div>
                        <div>
                            <span className="text-gray-500">Создан:</span>
                            <span className="ml-2">{new Date(schedule.created_at).toLocaleDateString('ru-RU')}</span>
                        </div>

                        {schedule.interval_unit && (
                            <div>
                                <span className="text-gray-500">Периодичность:</span>
                                <span className="ml-2 font-medium">
                                    {schedule.interval_value} {INTERVAL_LABELS[schedule.interval_unit] || schedule.interval_unit}
                                </span>
                            </div>
                        )}

                        {schedule.next_due_at && (
                            <div>
                                <span className="text-gray-500">Следующее ТО:</span>
                                <span className={`ml-2 font-medium ${overdue ? 'text-red-600' : 'text-gray-800'}`}>
                                    {overdue && '⚠ '}{new Date(schedule.next_due_at).toLocaleDateString('ru-RU')}
                                </span>
                            </div>
                        )}

                        {schedule.last_meter_value != null && (
                            <div>
                                <span className="text-gray-500">Последнее показание счётчика:</span>
                                <span className="ml-2 font-medium">{schedule.last_meter_value}</span>
                            </div>
                        )}

                        {schedule.description && (
                            <div className="col-span-2">
                                <span className="text-gray-500">Описание:</span>
                                <span className="ml-2 whitespace-pre-wrap">{schedule.description}</span>
                            </div>
                        )}
                    </div>
                )}

                {canManage && schedule.status !== 'cancelled' && schedule.status !== 'completed' && !editing && (
                    <div className="flex gap-3 pt-2 border-t border-gray-100">
                        <button onClick={handleDeactivate} disabled={updating}
                            className="px-4 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-60">
                            Деактивировать
                        </button>
                        <button onClick={handleDelete}
                            className="px-4 py-2 border border-red-300 text-red-600 rounded-lg text-sm hover:bg-red-50">
                            Удалить
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
