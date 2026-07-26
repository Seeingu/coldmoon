// async function* foo() {
//     yield 1;
//     yield 2;
// }
//
// (async function () {
//     for await (const num of foo()) {
//         console.log(num);
//         // Expected output: 1
//
//         break; // Closes iterator, triggers return
//     }
// })();

const foo = function* () {
    yield 'a';
    yield 'b';
    yield 'c';
};

let str = '';
for (const val of foo()) {
    str = str + val;
}
console.log( str)
