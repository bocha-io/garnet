// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {System} from "@latticexyz/world/src/System.sol";
import {getKeysWithValue} from "@latticexyz/world/src/modules/keyswithvalue/getKeysWithValue.sol";
import {User, UserTableId} from "../codegen/tables/User.sol";
import {PlayerOne} from "../codegen/tables/PlayerOne.sol";
import {addressToEntityKey} from "../addressToEntityKey.sol";

contract RegisterSystem is System {
    function register(bytes32 name) public {
        bytes32 senderKey = addressToEntityKey(_msgSender());
        require(User.get(senderKey) == 0, "wallet already registered");
        bytes32[] memory userKeys = getKeysWithValue(UserTableId, User.encode(name));
        require(userKeys.length == 0, "the username is already registered");
        User.set(senderKey, name);
    }
}
