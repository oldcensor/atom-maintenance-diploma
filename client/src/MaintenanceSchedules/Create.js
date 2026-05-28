import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
    useGetEquipmentQuery,
    useCreateMaintenanceScheduleMutation,
} from '../store/apiSlice';
import { toast } from 'react-toastify';
import Config from '../utils/Config';

const METER_LABELS = {
    operating_hours: 'моточасов',
    cycles: 'циклов',
    days: 'дней',
};

export default function ScheduleCreate() {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();

    const { data: equipment } = useGetEquipmentQuery();
    const [createSchedule, { isLoading: isPending }] = useCreateMaintenanceScheduleMutation();

    const [form, setForm] = useState({
        equipment_id: searchParams.get('equipment_id') || '',
        scheduled_at: new Date().toISOString().slice(0, 16),
        description: '',
        interval_unit: '',
        interval_value: '',
    });

    const [telemetry, setTelemetry] = useState(null);

    useEffect(() => {
        if (!form.equipment_id) { setTelemetry(null); return; }
        fetch(`${Config.endpoints.simulatorUrl}/api/v1/telemetry/${form.equipment_id}`)
            .then(r => r.ok ? r.json() : null)
            .then(data => setTelemetry(data))
            .catch(() => setTelemetry(null));
    }, [form.equipment_id]);

    const set = (field) => (e) => setForm(f => ({ ...f, [field]: e.target.value }));

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            const body = {
                equipment_id: parseInt(form.equipment_id, 10),
                scheduled_at: new Date(form.scheduled_at).toISOString(),
                description: form.description,
                assigned_to: null,
            };
            if (form.interval_unit) {
                body.interval_unit = form.interval_unit;
                body.interval_value = parseInt(form.interval_value, 10);
            }
            await createSchedule(body).unwrap();
            toast.success('Регламент создан');
            navigate('/schedules');
        } catch (err) {
            toast.error(err.data?.error || err.message || 'Ошибка при создании');
        }
    };

    return (
        <div className="max-w-2xl">
            <div className="flex items-center gap-3 mb-6">
                <button onClick={() => navigate('/schedules')} className="text-gray-400 hover:text-gray-600">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
                    </svg>
                </button>
                <h1 className="text-2xl font-bold text-gray-800">Новый регламент</h1>
            </div>

            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="col-span-2">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Оборудование *</label>
                            <select required value={form.equipment_id} onChange={set('equipment_id')}
                                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                                <option value="">Выберите оборудование</option>
                                {(equipment || []).filter(e => e.status !== 'decommissioned').map(e => (
                                    <option key={e.id} value={e.id}>{e.name} ({e.serial_number})</option>
                                ))}
                            </select>
                        </div>

                        {telemetry && (
                            <div className="col-span-2 flex items-center gap-3 bg-blue-50 border border-blue-200 rounded-lg px-4 py-2.5 text-sm">
                                <span className="text-blue-400">
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 0 1 3 19.875v-6.75ZM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V8.625ZM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V4.125Z" />
                                    </svg>
                                </span>
                                <span className="text-blue-700">
                                    Текущая наработка симулятора:&nbsp;
                                    <span className="font-bold tabular-nums">
                                        {telemetry.current_value.toLocaleString('ru-RU')}
                                    </span>
                                    &nbsp;{METER_LABELS[telemetry.meter_type] || telemetry.unit}
                                </span>
                            </div>
                        )}

                        <div className="col-span-2">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Дата планирования *</label>
                            <input type="datetime-local" required value={form.scheduled_at} onChange={set('scheduled_at')}
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
                                        <input
                                            type="radio"
                                            name="interval_unit"
                                            value={opt.value}
                                            checked={form.interval_unit === opt.value}
                                            onChange={set('interval_unit')}
                                            className="accent-blue-700"
                                        />
                                        {opt.label}
                                    </label>
                                ))}
                            </div>
                        </div>

                        {form.interval_unit && (
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Значение интервала *</label>
                                <input
                                    type="number"
                                    required
                                    min={1}
                                    value={form.interval_value}
                                    onChange={set('interval_value')}
                                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    placeholder="90"
                                />
                            </div>
                        )}

                        <div className="col-span-2">
                            <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
                            <textarea rows={3} maxLength={1000} value={form.description} onChange={set('description')}
                                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                                placeholder="Описание регламентных работ" />
                        </div>
                    </div>

                    <div className="flex gap-3 pt-2">
                        <button type="submit" disabled={isPending}
                            className="bg-blue-800 text-white px-6 py-2 rounded-lg text-sm hover:bg-blue-900 disabled:opacity-60 transition-colors">
                            {isPending ? 'Создание...' : 'Создать'}
                        </button>
                        <button type="button" onClick={() => navigate('/schedules')}
                            className="px-6 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50">
                            Отмена
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}
