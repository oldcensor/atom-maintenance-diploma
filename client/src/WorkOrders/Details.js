import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    useGetWorkOrderByIdQuery,
    useGetEquipmentQuery,
    useGetEmployeesQuery,
    useGetInspectionReportsQuery,
    useUpdateWorkOrderMutation,
    useGetStatusLogQuery,
    useGetWOCommentsQuery,
    useCreateWOCommentMutation,
    useDeleteWOCommentMutation,
    useGetChecklistQuery,
    useCreateChecklistItemMutation,
    useToggleChecklistItemMutation,
    useDeleteChecklistItemMutation,
} from '../store/apiSlice';
import { useAuth, hasAtLeastRole } from '../auth/AuthContext';
import Loading from '../utils/Loading';
import { toast } from 'react-toastify';
import InspectionReportCreate from '../InspectionReports/Create';

const ROLE_LABELS = { technician: 'Техник', engineer: 'Инженер', manager: 'Менеджер', admin: 'Администратор' };

const STATUS_LABELS = { open: 'Открыт', in_progress: 'В работе', completed: 'Завершён', cancelled: 'Отменён' };

const WO_TYPE_LABELS = { emergency: 'Аварийный', corrective: 'Корректирующий', planned: 'Плановый' };
const WO_TYPE_COLORS = {
    emergency:  'bg-red-100 text-red-700',
    corrective: 'bg-yellow-100 text-yellow-700',
    planned:    'bg-blue-100 text-blue-700',
};

export default function WorkOrderDetails() {
    const { id } = useParams();
    const navigate = useNavigate();
    const { user } = useAuth();

    // Права: менеджер+ видит список сотрудников
    const isAtLeastManager = hasAtLeastRole(user?.role, 'manager');
    const isAtLeastEngineer = hasAtLeastRole(user?.role, 'engineer');

    const { data: wo, isLoading, error } = useGetWorkOrderByIdQuery(id);
    const { data: equipment } = useGetEquipmentQuery();
    const { data: employees } = useGetEmployeesQuery();
    const { data: reports } = useGetInspectionReportsQuery();
    const [updateWo, { isLoading: updating }] = useUpdateWorkOrderMutation();

    // История статусов
    const { data: statusLog = [] } = useGetStatusLogQuery(id);

    // Комментарии
    const { data: comments = [] } = useGetWOCommentsQuery(id);
    const [createComment] = useCreateWOCommentMutation();
    const [deleteComment] = useDeleteWOCommentMutation();
    const [commentText, setCommentText] = useState('');

    // Чек-лист
    const { data: checklist = [] } = useGetChecklistQuery(id);
    const [createChecklistItem] = useCreateChecklistItemMutation();
    const [toggleChecklistItem] = useToggleChecklistItemMutation();
    const [deleteChecklistItem] = useDeleteChecklistItemMutation();
    const [checklistText, setChecklistText] = useState('');

    const checklistDone = checklist.filter(i => i.checked).length;
    const checklistAllChecked = checklist.length > 0 && checklistDone === checklist.length;

    const [showReportForm, setShowReportForm] = useState(false);
    const [showReassign, setShowReassign] = useState(false);
    const [newAssignee, setNewAssignee] = useState('');

    const eqMap = Object.fromEntries((equipment || []).map(e => [e.id, e.name]));
    const empMap = Object.fromEntries((employees || []).map(e => [e.id, e.full_name]));

    const report = (reports || []).find(r => r.work_order_id === parseInt(id, 10));

    const changeStatus = async (newStatus) => {
        if (newStatus === 'completed' && checklist.length > 0 && !checklistAllChecked) {
            toast.warning('Для завершения наряда необходимо отметить все пункты чек-листа');
            return;
        }
        if (newStatus === 'completed' && !report) {
            toast.warning('Для завершения наряда необходимо сначала создать протокол');
            return;
        }
        // При взятии в работу незанятого наряда — назначаем себя исполнителем
        const assignedTo = (newStatus === 'in_progress' && !wo.assigned_to)
            ? user?.id
            : (wo.assigned_to || null);
        try {
            await updateWo({
                id,
                equipment_id: wo.equipment_id,
                title: wo.title,
                description: wo.description || '',
                assigned_to: assignedTo,
                schedule_id: wo.schedule_id || null,
                status: newStatus,
                work_type: wo.work_type || 'corrective',
                completed_at: newStatus === 'completed' ? new Date().toISOString() : null,
            }).unwrap();
            toast.success('Статус обновлён');
        } catch (err) {
            toast.error(err.data?.error || err.message || 'Ошибка');
        }
    };

    const handleReassign = async (e) => {
        e.preventDefault();
        try {
            await updateWo({
                id,
                equipment_id: wo.equipment_id,
                title: wo.title,
                description: wo.description || '',
                assigned_to: newAssignee ? parseInt(newAssignee, 10) : null,
                schedule_id: wo.schedule_id || null,
                status: wo.status,
                work_type: wo.work_type || 'corrective',
                completed_at: wo.completed_at || null,
            }).unwrap();
            toast.success('Исполнитель переназначен');
            setShowReassign(false);
        } catch (err) {
            toast.error(err.data?.error || err.message || 'Ошибка');
        }
    };

    if (isLoading) return <div className="flex justify-center py-12"><Loading /></div>;
    if (error) return <div className="text-red-500 p-4">{error.data?.error || 'Ошибка загрузки'}</div>;
    if (!wo) return null;

    const isActive = wo.status === 'open' || wo.status === 'in_progress';
    const isAssignedToMe = wo.assigned_to === user?.id;
    const isUnassigned = !wo.assigned_to;

    // Взять в работу / завершить: инженер+ или исполнитель наряда (или незанятый наряд)
    const canActOnWorkOrder = isAtLeastEngineer || isAssignedToMe || isUnassigned;
    // Отменить / переназначить: инженер+
    const canManage = isAtLeastEngineer;

    return (
        <div className="space-y-6 max-w-3xl">
            <div className="flex items-center gap-3">
                <button onClick={() => navigate('/work-orders')} className="text-gray-400 hover:text-gray-600">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
                    </svg>
                </button>
                <h1 className="text-xl font-bold text-gray-800 flex-1">{wo.title}</h1>
                {wo.work_type && (
                    <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${WO_TYPE_COLORS[wo.work_type] || 'bg-gray-100 text-gray-500'}`}>{WO_TYPE_LABELS[wo.work_type] || wo.work_type}</span>
                )}
                <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${
                    wo.status === 'open' ? 'bg-blue-100 text-blue-700' :
                    wo.status === 'in_progress' ? 'bg-yellow-100 text-yellow-700' :
                    wo.status === 'completed' ? 'bg-green-100 text-green-700' :
                    'bg-gray-100 text-gray-500'
                }`}>
                    {STATUS_LABELS[wo.status] || wo.status}
                </span>
            </div>

            {/* Основная информация */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 space-y-4">
                <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                        <span className="text-gray-500">Оборудование:</span>
                        <button className="ml-2 text-blue-600 hover:underline" onClick={() => navigate(`/equipment/${wo.equipment_id}`)}>
                            {eqMap[wo.equipment_id] || `#${wo.equipment_id}`}
                        </button>
                    </div>
                    {wo.schedule_id && (
                        <div>
                            <span className="text-gray-500">Регламент:</span>
                            <button className="ml-2 text-blue-600 hover:underline" onClick={() => navigate(`/schedules/${wo.schedule_id}`)}>
                                #{wo.schedule_id}
                            </button>
                        </div>
                    )}
                    <div className="flex items-center gap-2">
                        <span className="text-gray-500">Исполнитель:</span>
                        <span className="ml-2">{wo.assigned_to ? empMap[wo.assigned_to] || `#${wo.assigned_to}` : '—'}</span>
                        {isAtLeastManager && isActive && (
                            <button onClick={() => { setShowReassign(!showReassign); setNewAssignee(wo.assigned_to || ''); }}
                                className="text-xs text-blue-600 hover:text-blue-800 ml-1">
                                [изменить]
                            </button>
                        )}
                    </div>
                    {wo.created_by && (
                        <div>
                            <span className="text-gray-500">Поставил задачу:</span>
                            <span className="ml-2">{empMap[wo.created_by] || `#${wo.created_by}`}</span>
                        </div>
                    )}
                    <div>
                        <span className="text-gray-500">Создан:</span>
                        <span className="ml-2">{new Date(wo.created_at).toLocaleDateString('ru-RU')}</span>
                    </div>
                    {wo.completed_at && (
                        <div>
                            <span className="text-gray-500">Завершён:</span>
                            <span className="ml-2">{new Date(wo.completed_at).toLocaleDateString('ru-RU')}</span>
                        </div>
                    )}
                    {wo.description && (
                        <div className="col-span-2">
                            <span className="text-gray-500">Описание:</span>
                            <span className="ml-2 whitespace-pre-wrap">{wo.description}</span>
                        </div>
                    )}
                </div>

                {/* Форма переназначения — только менеджер+ (у них есть список сотрудников) */}
                {showReassign && isAtLeastManager && (
                    <form onSubmit={handleReassign} className="flex gap-3 items-end pt-2 border-t border-gray-100">
                        <div className="flex-1">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Новый исполнитель</label>
                            <select value={newAssignee} onChange={e => setNewAssignee(e.target.value)}
                                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                <option value="">Не назначен</option>
                                {(employees || []).filter(emp => ['technician', 'engineer'].includes(emp.role)).map(emp => <option key={emp.id} value={emp.id}>{emp.full_name} ({ROLE_LABELS[emp.role] ?? emp.role})</option>)}
                            </select>
                        </div>
                        <button type="submit" disabled={updating}
                            className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900 disabled:opacity-60">
                            Сохранить
                        </button>
                        <button type="button" onClick={() => setShowReassign(false)}
                            className="px-4 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50">
                            Отмена
                        </button>
                    </form>
                )}

                {/* Переходы статуса */}
                {(canActOnWorkOrder || canManage) && (
                    <div className="flex gap-3 pt-2 border-t border-gray-100 flex-wrap">
                        {wo.status === 'open' && canActOnWorkOrder && (
                            <button onClick={() => changeStatus('in_progress')} disabled={updating}
                                className="bg-yellow-500 text-white px-4 py-2 rounded-lg text-sm hover:bg-yellow-600 disabled:opacity-60">
                                Взять в работу
                            </button>
                        )}
                        {wo.status === 'in_progress' && canActOnWorkOrder && (
                            <button onClick={() => changeStatus('completed')} disabled={updating}
                                className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-700 disabled:opacity-60">
                                Завершить
                            </button>
                        )}
                        {(wo.status === 'open' || wo.status === 'in_progress') && canManage && (
                            <button onClick={() => changeStatus('cancelled')} disabled={updating}
                                className="px-4 py-2 border border-red-300 text-red-600 rounded-lg text-sm hover:bg-red-50 disabled:opacity-60">
                                Отменить
                            </button>
                        )}
                    </div>
                )}
            </div>

            {/* Чек-лист */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
                    <h2 className="font-semibold text-gray-800">Чек-лист</h2>
                    {checklist.length > 0 && (
                        <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${
                            checklistAllChecked ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
                        }`}>
                            {checklistDone} / {checklist.length}
                        </span>
                    )}
                </div>

                {checklist.length === 0 ? (
                    <div className="px-5 py-6 text-center text-gray-400 text-sm">Пунктов нет</div>
                ) : (
                    <div className="divide-y divide-gray-100">
                        {checklist.map(item => (
                            <div key={item.id} className="px-5 py-3 flex items-start gap-3">
                                <input
                                    type="checkbox"
                                    checked={item.checked}
                                    disabled={!isActive || updating}
                                    onChange={async () => {
                                        try {
                                            await toggleChecklistItem({ woId: id, itemId: item.id, checked: !item.checked }).unwrap();
                                        } catch (err) {
                                            toast.error(err.data?.message || err.data?.error || 'Ошибка');
                                        }
                                    }}
                                    className="mt-0.5 h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500 disabled:opacity-50 cursor-pointer disabled:cursor-default"
                                />
                                <div className="flex-1 min-w-0">
                                    <p className={`text-sm break-words ${item.checked ? 'text-gray-400 line-through' : 'text-gray-800'}`}>
                                        {item.text}
                                    </p>
                                    {item.checked && item.checked_by && (
                                        <p className="text-xs text-gray-400 mt-0.5">
                                            {empMap[item.checked_by] || `#${item.checked_by}`}
                                            {item.checked_at && ` · ${new Date(item.checked_at).toLocaleString('ru-RU')}`}
                                        </p>
                                    )}
                                </div>
                                {isAtLeastEngineer && isActive && (
                                    <button
                                        onClick={async () => {
                                            try {
                                                await deleteChecklistItem({ woId: id, itemId: item.id }).unwrap();
                                            } catch (err) {
                                                toast.error(err.data?.message || err.data?.error || 'Ошибка');
                                            }
                                        }}
                                        className="text-gray-300 hover:text-red-500 transition-colors flex-shrink-0"
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                            <path strokeLinecap="round" strokeLinejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
                                        </svg>
                                    </button>
                                )}
                            </div>
                        ))}
                    </div>
                )}

                {/* Добавление пункта — инженер+ и только для активного наряда */}
                {isAtLeastEngineer && isActive && (
                    <form
                        onSubmit={async (e) => {
                            e.preventDefault();
                            if (!checklistText.trim()) return;
                            try {
                                await createChecklistItem({ woId: id, text: checklistText.trim(), sort_order: checklist.length }).unwrap();
                                setChecklistText('');
                            } catch (err) {
                                toast.error(err.data?.message || err.data?.error || 'Ошибка');
                            }
                        }}
                        className="flex gap-2 px-5 py-3 border-t border-gray-100"
                    >
                        <input
                            value={checklistText}
                            onChange={e => setChecklistText(e.target.value)}
                            maxLength={500}
                            placeholder="Добавить пункт..."
                            className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        <button type="submit" className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900">
                            Добавить
                        </button>
                    </form>
                )}
            </div>

            {/* Протокол выполнения */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
                    <h2 className="font-semibold text-gray-800">Протокол выполнения</h2>
                    {!report && wo.status === 'in_progress' && (canActOnWorkOrder) && (
                        <button onClick={() => setShowReportForm(!showReportForm)}
                            className="text-sm text-blue-600 hover:text-blue-800">
                            {showReportForm ? 'Скрыть' : '+ Создать протокол'}
                        </button>
                    )}
                </div>

                {showReportForm && (
                    <div className="p-5 border-b border-gray-100">
                        <InspectionReportCreate
                            workOrderId={parseInt(id, 10)}
                            inspectorId={user?.id}
                            onSuccess={() => { setShowReportForm(false); toast.success('Протокол создан'); }}
                        />
                    </div>
                )}

                {report ? (
                    <div className="p-5 space-y-3 text-sm">
                        <div>
                            <span className="text-gray-500">Инспектор:</span>
                            <span className="ml-2">{empMap[report.inspector_id] || `#${report.inspector_id}`}</span>
                        </div>
                        <div>
                            <span className="text-gray-500">Дата:</span>
                            <span className="ml-2">{new Date(report.created_at).toLocaleDateString('ru-RU')}</span>
                        </div>
                        <div>
                            <p className="text-gray-500 mb-1">Выявленные отклонения:</p>
                            <p className="bg-gray-50 rounded-lg p-3 text-gray-800">{report.findings}</p>
                        </div>
                        {report.recommendations && (
                            <div>
                                <p className="text-gray-500 mb-1">Рекомендации:</p>
                                <p className="bg-gray-50 rounded-lg p-3 text-gray-800">{report.recommendations}</p>
                            </div>
                        )}
                    </div>
                ) : (
                    !showReportForm && (
                        <div className="px-5 py-8 text-center text-gray-400 text-sm">
                            {wo.status === 'in_progress'
                                ? 'Протокол ещё не создан. Создайте протокол для завершения наряда.'
                                : 'Протокол не создан'}
                        </div>
                    )
                )}
            </div>

            {/* История статусов */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="px-5 py-4 border-b border-gray-100">
                    <h2 className="font-semibold text-gray-800">История изменений</h2>
                </div>
                {statusLog.length === 0 ? (
                    <div className="px-5 py-6 text-center text-gray-400 text-sm">Нет записей</div>
                ) : (
                    <div className="px-5 py-4">
                        <div className="relative border-l-2 border-gray-200 pl-6 space-y-4">
                            {statusLog.map(entry => {
                                const dotColor =
                                    entry.to_status === 'completed' ? 'bg-green-500' :
                                    entry.to_status === 'in_progress' ? 'bg-yellow-500' :
                                    entry.to_status === 'cancelled' ? 'bg-red-500' : 'bg-blue-500';
                                return (
                                    <div key={entry.id} className="relative">
                                        <div className={`absolute -left-[31px] top-1 w-3 h-3 rounded-full ${dotColor} ring-2 ring-white`} />
                                        <div className="text-sm">
                                            <span className="text-gray-500">
                                                {STATUS_LABELS[entry.from_status] || entry.from_status}
                                            </span>
                                            <span className="mx-1 text-gray-400">&rarr;</span>
                                            <span className="font-medium text-gray-800">
                                                {STATUS_LABELS[entry.to_status] || entry.to_status}
                                            </span>
                                            {entry.changed_by && (
                                                <span className="ml-2 text-gray-400">
                                                    {empMap[entry.changed_by] || `#${entry.changed_by}`}
                                                </span>
                                            )}
                                        </div>
                                        <div className="text-xs text-gray-400 mt-0.5">
                                            {new Date(entry.created_at).toLocaleString('ru-RU')}
                                        </div>
                                        {entry.comment && (
                                            <div className="text-sm text-gray-600 mt-1 bg-gray-50 rounded px-2 py-1">
                                                {entry.comment}
                                            </div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                )}
            </div>

            {/* Комментарии */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="px-5 py-4 border-b border-gray-100">
                    <h2 className="font-semibold text-gray-800">Комментарии</h2>
                </div>
                {comments.length === 0 ? (
                    <div className="px-5 py-6 text-center text-gray-400 text-sm">Нет комментариев</div>
                ) : (
                    <div className="divide-y divide-gray-100">
                        {comments.map(c => (
                            <div key={c.id} className="px-5 py-3 flex gap-3">
                                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center text-xs font-bold">
                                    {(empMap[c.author_id] || '?').charAt(0).toUpperCase()}
                                </div>
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center gap-2">
                                        <span className="text-sm font-medium text-gray-800">{empMap[c.author_id] || `#${c.author_id}`}</span>
                                        <span className="text-xs text-gray-400">{new Date(c.created_at).toLocaleString('ru-RU')}</span>
                                    </div>
                                    <p className="text-sm text-gray-600 mt-1 whitespace-pre-wrap break-words">{c.text}</p>
                                </div>
                                {c.author_id === user?.id && (
                                    <button
                                        onClick={async () => {
                                            try {
                                                await deleteComment({ woId: id, id: c.id }).unwrap();
                                            } catch (err) {
                                                toast.error(err.data?.error || 'Ошибка');
                                            }
                                        }}
                                        className="text-gray-300 hover:text-red-500 transition-colors flex-shrink-0"
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                            <path strokeLinecap="round" strokeLinejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
                                        </svg>
                                    </button>
                                )}
                            </div>
                        ))}
                    </div>
                )}

                <form
                    onSubmit={async (e) => {
                        e.preventDefault();
                        if (!commentText.trim()) return;
                        try {
                            await createComment({ woId: id, text: commentText.trim() }).unwrap();
                            setCommentText('');
                        } catch (err) {
                            toast.error(err.data?.error || 'Ошибка');
                        }
                    }}
                    className="flex gap-2 px-5 py-3 border-t border-gray-100"
                >
                    <input
                        value={commentText}
                        onChange={e => setCommentText(e.target.value)}
                        maxLength={500}
                        placeholder="Написать комментарий..."
                        className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <button type="submit" className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-900">
                        Отправить
                    </button>
                </form>
            </div>
        </div>
    );
}
