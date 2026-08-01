package tests

import "testing"

func TestPrivateMethodsAccessorsFieldsAndOptionalAccess(t *testing.T) {
	testSource(t, `
class Secret {
    #value;

    constructor(value) {
        this.#value = value;
    }

    #double() {
        return this.#value * 2;
    }

    get #current() {
        return this.#value;
    }

    set #current(value) {
        this.#value = value;
    }

    read() {
        return this.#double();
    }

    update(value) {
        this.#current = value;
        return this.#current;
    }

    inspect(other) {
        return other?.#value;
    }
}

const secret = new Secret(6);
assertEqual(secret.read(), 12, "private method");
assertEqual(secret.update(9), 9, "private accessor pair");
assertEqual(secret.inspect(secret), 9, "optional private access");
assertEqual(secret.inspect(null), undefined, "optional private short circuit");

let brandCheckThrew = false;
try {
    secret.inspect({});
} catch (error) {
    brandCheckThrew = true;
}
assertEqual(brandCheckThrew, true, "private brand check");
`)
}

func TestStaticPrivateMethod(t *testing.T) {
	testSource(t, `
class Answer {
    static #value() { return 42; }
    static read() { return this.#value(); }
}
assertEqual(Answer.read(), 42);
`)
}

func TestPublicGetterAndSetterDefinitionsRemainPaired(t *testing.T) {
	testSource(t, `
const object = {
    get value() { return this.stored; },
    set value(next) { this.stored = next; }
};
object.value = 17;
assertEqual(object.value, 17);

class Box {
    get value() { return this.stored; }
    set value(next) { this.stored = next; }
}
const box = new Box();
box.value = 23;
assertEqual(box.value, 23);
`)
}
