import React from 'react';

const ConfirmDeleteModal = ({ onConfirm, onCancel, text }) => {
    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            <div className="border border-[#134074] bg-white rounded-lg p-6 w-full max-w-sm shadow-xl text-center">
                <h2 className="text-lg font-semibold text-gray-800 mb-4">Подтвердить действие</h2>
                <p className="text-gray-600 mb-6">{text}</p>
                <div className="flex justify-center space-x-4">
                    <button
                        onClick={() => onConfirm(true)} // Возвращаем true при подтверждении
                        className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600 transition"
                    >
                        Удалить
                    </button>
                    <button
                        onClick={() => onConfirm(false)} // Возвращаем false при отмене
                        className="bg-gray-200 text-gray-800 px-4 py-2 rounded hover:bg-gray-300 transition"
                    >
                        Отменить
                    </button>
                </div>
            </div>
        </div>
    );
};

export default ConfirmDeleteModal;