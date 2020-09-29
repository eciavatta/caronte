
const validation = {
    isValidColor: (color) => true, // TODO
    isValidPort: (port, required) => parseInt(port) > (required ? 0 : -1) && parseInt(port) <= 65565,
    isValidAddress: (address) => true // TODO
};

export default validation;
