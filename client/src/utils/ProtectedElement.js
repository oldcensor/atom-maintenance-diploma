import React from 'react';
import { hasAtLeastRole } from '../auth/AuthContext';

// requiredRole — минимальная роль для отображения элемента
// userRole    — роль текущего пользователя
const ProtectedElement = ({ userRole, requiredRole, children }) => {
    if (!hasAtLeastRole(userRole, requiredRole)) return null;
    return <>{children}</>;
};

export default ProtectedElement;
