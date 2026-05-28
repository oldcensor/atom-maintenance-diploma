import { createSlice } from '@reduxjs/toolkit';

const authSlice = createSlice({
    name: 'auth',
    initialState: { token: null },
    reducers: {
        tokenSet: (state, { payload }) => { state.token = payload; },
        tokenCleared: (state) => { state.token = null; },
    },
});

export const { tokenSet, tokenCleared } = authSlice.actions;
export default authSlice.reducer;
