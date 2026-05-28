const Config = {
    endpoints: {
        baseUrl: process.env.REACT_APP_API_URL || 'http://localhost:8080',
        simulatorUrl: process.env.REACT_APP_SIMULATOR_URL || 'http://localhost:8090',
    },
};

export default Config;
