import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './auth/AuthContext';
import Navbar from './Navbar';
import Loading from './utils/Loading';
import Login from './auth/Login';
import Home from './Home';
import EquipmentList from './Equipment/List';
import EquipmentDetails from './Equipment/Details';
import EquipmentCreate from './Equipment/Create';
import ScheduleList from './MaintenanceSchedules/List';
import ScheduleCreate from './MaintenanceSchedules/Create';
import ScheduleDetails from './MaintenanceSchedules/Details';
import WorkOrderList from './WorkOrders/List';
import WorkOrderCreate from './WorkOrders/Create';
import WorkOrderDetails from './WorkOrders/Details';
import EmployeeList from './Employees/List';
import DepartmentList from './Departments/List';
import EquipmentTypeList from './EquipmentTypes/List';

const ProtectedRoute = ({ element }) => {
    const { isAuthenticated, loadingAuth } = useAuth();
    if (loadingAuth) {
        return (
            <div className="flex justify-center items-center h-screen bg-gray-100">
                <Loading />
            </div>
        );
    }
    if (!isAuthenticated) return <Navigate to="/login" replace />;
    return element;
};

const Layout = ({ children }) => (
    <>
        <Navbar />
        <div className="min-h-screen bg-gray-100 pb-8">
            <div className="max-w-screen-xl mx-auto px-4 py-4">
                {children}
            </div>
        </div>
    </>
);

function App() {
    return (
        <AuthProvider>
            <Router>
                <Routes>
                    <Route path="/login" element={<Login />} />
                    <Route path="/*" element={
                        <ProtectedRoute element={
                            <Layout>
                                <Routes>
                                    <Route path="/" element={<Home />} />
                                    <Route path="/equipment" element={<EquipmentList />} />
                                    <Route path="/equipment/create" element={<EquipmentCreate />} />
                                    <Route path="/equipment/:id" element={<EquipmentDetails />} />
                                    <Route path="/schedules" element={<ScheduleList />} />
                                    <Route path="/schedules/create" element={<ScheduleCreate />} />
                                    <Route path="/schedules/:id" element={<ScheduleDetails />} />
                                    <Route path="/work-orders" element={<WorkOrderList />} />
                                    <Route path="/work-orders/create" element={<WorkOrderCreate />} />
                                    <Route path="/work-orders/:id" element={<WorkOrderDetails />} />
                                    <Route path="/employees" element={<EmployeeList />} />
                                    <Route path="/departments" element={<DepartmentList />} />
                                    <Route path="/equipment-types" element={<EquipmentTypeList />} />
                                    <Route path="*" element={<Navigate to="/" replace />} />
                                </Routes>
                            </Layout>
                        } />
                    } />
                </Routes>
            </Router>
        </AuthProvider>
    );
}

export default App;
