
const validation = {
    isValidColor: (color) => /^#(?:[0-9a-fA-F]{3}){1,2}$/.test(color),
    isValidPort: (port, required) => parseInt(port) > (required ? 0 : -1) && parseInt(port) <= 65565,
    isValidAddress: (address) => true // TODO
};

export default validation;
