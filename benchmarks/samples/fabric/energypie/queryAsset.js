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

module.exports.info = 'Querying assets.';

const helper = require('./helper');

let txIndex = 0;
let limitIndex, bc, contx;
const dic = 'abcdefghijklmnopqrstuvwxyz';


module.exports.init = async function(blockchain, context, args) {
    bc = blockchain;
    contx = context;
    // limitIndex = args.assets;

    // await helper.createAsset(bc, contx, args);

    return Promise.resolve();
};

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

module.exports.run = function() {
    txIndex++;
    let assetId = 'Client' + contx.clientIdx + get26Num(process.pid) + txIndex.toString();

    let args = {
        chaincodeFunction: 'queryAsset',
        chaincodeArguments: [assetId]
    };

    console.log(args);

    let Nspace = txIndex % 2 === 0 ? 'myasset' : 'yourasset';

    return bc.bcObj.querySmartContract(contx, Nspace, 'v1', args, 30);
};

module.exports.end = function() {
    return Promise.resolve();
};
