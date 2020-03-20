/*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

'use strict';

let types = ['aud', 'kwh', 'rec', 'goo', 'cc', 'srec', 'trc', 'ppa', 'usd','bitcoin'];
let values = ['1000', '10000', '20', '30', '10', '12', '20', '22', '500', '10'];
let owners = ['Tomoko', 'Brad', 'Jin Soo', 'Max', 'Adrianna', 'Michel', 'Aarav', 'Pari', 'Valeria', 'Shotaro'];

let Id;
let txIndex = 0;
const dic = 'abcdefghijklmnopqrstuvwxyz';

/**
 * Generate string by picking characters from dic variable
 * @param {*} number character to select
 * @returns {String} string generated based on @param number
 */
function get26Num(number){
    let result = '';
    while(number > 0) {
        result += dic.charAt(number % 26);
        number = parseInt(number/26);
    }
    return result;
}

module.exports.createAsset = async function (bc, contx, args, type, value, owner) {

    while (txIndex < args.assets) {
        txIndex++;
        Id = 'Client' + contx.clientIdx + get26Num(process.pid) + txIndex.toString();
        type = types[Math.floor(Math.random() * types.length)];
        value = values[Math.floor(Math.random() * values.length)];
        owner = owners[Math.floor(Math.random() * owners.length)];
    
        let Nspace = txIndex % 2 === 0 ? 'asset' : 'asset';

        let myArgs = {
            chaincodeFunction: 'createAsset',
            chaincodeArguments: [JSON.stringify({Id, type, owner, value, Nspace})]
        };
    

        await bc.invokeSmartContract(contx, Nspace, 'v1', myArgs, 30);
    }

};
