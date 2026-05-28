import { useState } from 'react';
import { useCreateInspectionReportMutation } from '../store/apiSlice';
import { toast } from 'react-toastify';

export default function InspectionReportCreate({ workOrderId, inspectorId, onSuccess }) {
    const [createReport, { isLoading: isPending }] = useCreateInspectionReportMutation();

    const [form, setForm] = useState({ findings: '', recommendations: '' });
    const set = (field) => (e) => setForm(f => ({ ...f, [field]: e.target.value }));

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await createReport({
                work_order_id: workOrderId,
                inspector_id: inspectorId,
                findings: form.findings,
                recommendations: form.recommendations,
            }).unwrap();
            onSuccess && onSuccess();
        } catch (err) {
            toast.error(err.data?.error || err.message || 'Ошибка при создании протокола');
        }
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                    Выявленные отклонения *
                </label>
                <textarea
                    required
                    rows={4}
                    maxLength={2000}
                    value={form.findings}
                    onChange={set('findings')}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                    placeholder="Опишите выявленные отклонения, дефекты, нарушения параметров"
                />
            </div>

            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Рекомендации</label>
                <textarea
                    rows={3}
                    maxLength={2000}
                    value={form.recommendations}
                    onChange={set('recommendations')}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                    placeholder="Рекомендации по дальнейшей эксплуатации или необходимым работам"
                />
            </div>

            <div className="flex gap-3">
                <button type="submit" disabled={isPending}
                    className="bg-blue-800 text-white px-6 py-2 rounded-lg text-sm hover:bg-blue-900 disabled:opacity-60 transition-colors">
                    {isPending ? 'Сохранение...' : 'Сохранить протокол'}
                </button>
            </div>
        </form>
    );
}
