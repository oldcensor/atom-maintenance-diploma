import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import { tokenCleared } from './authSlice';
import { callRefresh } from './refresher';
import Config from '../utils/Config';

const BASE = Config.endpoints.baseUrl;

const rawBaseQuery = fetchBaseQuery({
    baseUrl: `${BASE}/api/v1`,
    prepareHeaders: (headers, { getState }) => {
        const token = getState().auth.token;
        if (token) headers.set('Authorization', `Bearer ${token}`);
        return headers;
    },
});

const baseQuery = async (args, api, extraOptions) => {
    let result = await rawBaseQuery(args, api, extraOptions);

    if (result.error?.status === 401) {
        try {
            // Делегируем refresh в AuthContext — там есть mutex против гонки
            await callRefresh();
            result = await rawBaseQuery(args, api, extraOptions);
        } catch {
            api.dispatch(tokenCleared());
            localStorage.removeItem('refreshToken');
        }
    }

    return result;
};

export const apiSlice = createApi({
    reducerPath: 'api',
    baseQuery,
    tagTypes: ['Equipment', 'EquipmentType', 'Department', 'Employee', 'Schedule', 'WorkOrder', 'InspectionReport', 'StatusLog', 'WOComment', 'Checklist'],
    endpoints: (builder) => ({

        // ── Equipment ────────────────────────────────────────────────
        getEquipment: builder.query({
            query: () => '/equipment',
            providesTags: ['Equipment'],
        }),
        getEquipmentById: builder.query({
            query: (id) => `/equipment/${id}`,
            providesTags: (r, e, id) => [{ type: 'Equipment', id }],
        }),
        createEquipment: builder.mutation({
            query: (body) => ({ url: '/equipment', method: 'POST', body }),
            invalidatesTags: ['Equipment'],
        }),
        updateEquipment: builder.mutation({
            query: ({ id, ...body }) => ({ url: `/equipment/${id}`, method: 'PUT', body }),
            invalidatesTags: (r, e, { id }) => ['Equipment', { type: 'Equipment', id }],
        }),
        deleteEquipment: builder.mutation({
            query: (id) => ({ url: `/equipment/${id}`, method: 'DELETE' }),
            invalidatesTags: ['Equipment'],
        }),

        // ── Equipment Types ──────────────────────────────────────────
        getEquipmentTypes: builder.query({
            query: () => '/equipment-types',
            providesTags: ['EquipmentType'],
        }),
        createEquipmentType: builder.mutation({
            query: (body) => ({ url: '/equipment-types', method: 'POST', body }),
            invalidatesTags: ['EquipmentType'],
        }),
        updateEquipmentType: builder.mutation({
            query: ({ id, ...body }) => ({ url: `/equipment-types/${id}`, method: 'PUT', body }),
            invalidatesTags: ['EquipmentType'],
        }),
        deleteEquipmentType: builder.mutation({
            query: (id) => ({ url: `/equipment-types/${id}`, method: 'DELETE' }),
            invalidatesTags: ['EquipmentType'],
        }),

        // ── Departments ──────────────────────────────────────────────
        getDepartments: builder.query({
            query: () => '/departments',
            providesTags: ['Department'],
        }),
        createDepartment: builder.mutation({
            query: (body) => ({ url: '/departments', method: 'POST', body }),
            invalidatesTags: ['Department'],
        }),
        updateDepartment: builder.mutation({
            query: ({ id, ...body }) => ({ url: `/departments/${id}`, method: 'PUT', body }),
            invalidatesTags: ['Department'],
        }),
        deleteDepartment: builder.mutation({
            query: (id) => ({ url: `/departments/${id}`, method: 'DELETE' }),
            invalidatesTags: ['Department'],
        }),

        // ── Employees ────────────────────────────────────────────────
        getEmployees: builder.query({
            query: () => '/employees',
            providesTags: ['Employee'],
        }),
        createEmployee: builder.mutation({
            query: (body) => ({ url: '/employees', method: 'POST', body }),
            invalidatesTags: ['Employee'],
        }),
        updateEmployee: builder.mutation({
            query: ({ id, ...body }) => ({ url: `/employees/${id}`, method: 'PUT', body }),
            invalidatesTags: ['Employee'],
        }),
        deleteEmployee: builder.mutation({
            query: (id) => ({ url: `/employees/${id}`, method: 'DELETE' }),
            invalidatesTags: ['Employee'],
        }),

        // ── Maintenance Schedules ────────────────────────────────────
        getMaintenanceSchedules: builder.query({
            query: () => '/maintenance-schedules',
            providesTags: ['Schedule'],
        }),
        getMaintenanceScheduleById: builder.query({
            query: (id) => `/maintenance-schedules/${id}`,
            providesTags: (r, e, id) => [{ type: 'Schedule', id }],
        }),
        createMaintenanceSchedule: builder.mutation({
            query: (body) => ({ url: '/maintenance-schedules', method: 'POST', body }),
            invalidatesTags: ['Schedule'],
        }),
        updateMaintenanceSchedule: builder.mutation({
            query: ({ id, ...body }) => ({ url: `/maintenance-schedules/${id}`, method: 'PUT', body }),
            invalidatesTags: (r, e, { id }) => ['Schedule', { type: 'Schedule', id }],
        }),
        deleteMaintenanceSchedule: builder.mutation({
            query: (id) => ({ url: `/maintenance-schedules/${id}`, method: 'DELETE' }),
            invalidatesTags: ['Schedule'],
        }),

        // ── Work Orders ──────────────────────────────────────────────
        getWorkOrders: builder.query({
            query: () => '/work-orders',
            providesTags: ['WorkOrder'],
        }),
        getWorkOrderById: builder.query({
            query: (id) => `/work-orders/${id}`,
            providesTags: (r, e, id) => [{ type: 'WorkOrder', id }],
        }),
        createWorkOrder: builder.mutation({
            query: (body) => ({ url: '/work-orders', method: 'POST', body }),
            invalidatesTags: ['WorkOrder'],
        }),
        updateWorkOrder: builder.mutation({
            query: ({ id, ...body }) => ({ url: `/work-orders/${id}`, method: 'PUT', body }),
            invalidatesTags: (r, e, { id }) => ['WorkOrder', { type: 'WorkOrder', id }, 'StatusLog'],
        }),
        deleteWorkOrder: builder.mutation({
            query: (id) => ({ url: `/work-orders/${id}`, method: 'DELETE' }),
            invalidatesTags: ['WorkOrder'],
        }),

        // ── Inspection Reports ───────────────────────────────────────
        getInspectionReports: builder.query({
            query: () => '/inspection-reports',
            providesTags: ['InspectionReport'],
        }),
        createInspectionReport: builder.mutation({
            query: (body) => ({ url: '/inspection-reports', method: 'POST', body }),
            invalidatesTags: ['InspectionReport'],
        }),
        deleteInspectionReport: builder.mutation({
            query: (id) => ({ url: `/inspection-reports/${id}`, method: 'DELETE' }),
            invalidatesTags: ['InspectionReport'],
        }),

        // ── Status Log ─────────────────────────────────────────────────
        getStatusLog: builder.query({
            query: (woId) => `/work-orders/${woId}/status-log`,
            providesTags: (r, e, woId) => [{ type: 'StatusLog', id: woId }],
        }),

        // ── Work Order Comments ────────────────────────────────────────
        getWOComments: builder.query({
            query: (woId) => `/work-orders/${woId}/comments`,
            providesTags: (r, e, woId) => [{ type: 'WOComment', id: woId }],
        }),
        createWOComment: builder.mutation({
            query: ({ woId, text }) => ({ url: `/work-orders/${woId}/comments`, method: 'POST', body: { text } }),
            invalidatesTags: (r, e, { woId }) => [{ type: 'WOComment', id: woId }],
        }),
        deleteWOComment: builder.mutation({
            query: ({ woId, id }) => ({ url: `/work-orders/${woId}/comments/${id}`, method: 'DELETE' }),
            invalidatesTags: (r, e, { woId }) => [{ type: 'WOComment', id: woId }],
        }),

        // ── Work Order Checklist ───────────────────────────────────────
        getChecklist: builder.query({
            query: (woId) => `/work-orders/${woId}/checklist`,
            providesTags: (r, e, woId) => [{ type: 'Checklist', id: woId }],
        }),
        createChecklistItem: builder.mutation({
            query: ({ woId, text, sort_order }) => ({ url: `/work-orders/${woId}/checklist`, method: 'POST', body: { text, sort_order } }),
            invalidatesTags: (r, e, { woId }) => [{ type: 'Checklist', id: woId }],
        }),
        toggleChecklistItem: builder.mutation({
            query: ({ woId, itemId, checked }) => ({ url: `/work-orders/${woId}/checklist/${itemId}`, method: 'PATCH', body: { checked } }),
            invalidatesTags: (r, e, { woId }) => [{ type: 'Checklist', id: woId }],
        }),
        deleteChecklistItem: builder.mutation({
            query: ({ woId, itemId }) => ({ url: `/work-orders/${woId}/checklist/${itemId}`, method: 'DELETE' }),
            invalidatesTags: (r, e, { woId }) => [{ type: 'Checklist', id: woId }],
        }),
    }),
});

export const {
    useGetEquipmentQuery,
    useGetEquipmentByIdQuery,
    useCreateEquipmentMutation,
    useUpdateEquipmentMutation,
    useDeleteEquipmentMutation,
    useGetEquipmentTypesQuery,
    useCreateEquipmentTypeMutation,
    useUpdateEquipmentTypeMutation,
    useDeleteEquipmentTypeMutation,
    useGetDepartmentsQuery,
    useCreateDepartmentMutation,
    useUpdateDepartmentMutation,
    useDeleteDepartmentMutation,
    useGetEmployeesQuery,
    useCreateEmployeeMutation,
    useUpdateEmployeeMutation,
    useDeleteEmployeeMutation,
    useGetMaintenanceSchedulesQuery,
    useGetMaintenanceScheduleByIdQuery,
    useCreateMaintenanceScheduleMutation,
    useUpdateMaintenanceScheduleMutation,
    useDeleteMaintenanceScheduleMutation,
    useGetWorkOrdersQuery,
    useGetWorkOrderByIdQuery,
    useCreateWorkOrderMutation,
    useUpdateWorkOrderMutation,
    useDeleteWorkOrderMutation,
    useGetInspectionReportsQuery,
    useCreateInspectionReportMutation,
    useDeleteInspectionReportMutation,
    useGetStatusLogQuery,
    useGetWOCommentsQuery,
    useCreateWOCommentMutation,
    useDeleteWOCommentMutation,
    useGetChecklistQuery,
    useCreateChecklistItemMutation,
    useToggleChecklistItemMutation,
    useDeleteChecklistItemMutation,
} = apiSlice;
