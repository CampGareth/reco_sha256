/*--------------------------------------------
  Example host code for SHA256
--------------------------------------------*/

package main

import (
	"crypto/sha256/host"
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/ReconfigureIO/sdaccel/xcl"
)

func reverseBytes(s []byte) []byte {
	r := make([]byte, 4)
	for i := 0; i < 4; i++ {
		r[i] = s[3-i]
	}
	return r
}

func main() {

	const (
		checkMark = "\u2713"
		ballotX   = "\u2717"
	)

	world := xcl.NewWorld()
	defer world.Release()
	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()

	// Pad message & calculate number of 64byte blocks

	// fox test string
	testString := "The quick brown fox jumps over the lazy dog."

	/*---
		// JaneEyre test string - only use to test a build, not in simulation

		testString := `Reader, I married him. A quiet wedding we had: he and I, the parson and clerk, were alone present. When we got back from church, I went into the
			kitchen of the manor-house, where Mary was cooking the dinner and John cleaning the knives, and I said-'Mary, I have been married to Mr. Rochester
			this morning.' The housekeeper and her husband were both of that decent phlegmatic order of people, to whom one may at any time safely communicate
			a remarkable piece of news without incurring the danger of having one's ears pierced by some shrill ejaculation, and subsequently stunned by a torrent
			of wordy wonderment. Mary did look up, and she did stare at me: the ladle with which she was basting a pair of chickens roasting at the fire, did for
			some three minutes hang suspended in air; and for the same space of time John's knives also had rest from the polishing process: but Mary, bending again
			over the roast, said only-'Have you, Miss? Well, for sure!' A short time after she pursued- 'I seed you go out with the master, but I didn't know you were
			gone to church to be wed;' and she basted away. John, when I turned to him, was grinning from ear to ear.'I telled Mary how it would be,' he said: 'I knew
			what Mr. Edward' (John was an old servant, and had known his master when he was the cadet of the house, therefore, he often gave him his Christian name)-
			'I knew what Mr. Edward would do; and I was certain he would not wait long neither: and he's done right, for aught I know. I wish you joy, Miss!' and he
			politely pulled his forelock. 'Thank you, John. Mr. Rochester told me to give you and Mary this.' I put into his hand a five-pound note. Without waiting
			to hear more, I left the kitchen. In passing the door of that sanctum some time after, I caught the words-'She'll happen do better for him nor ony o' t'
			grand ladies.' And again, 'If she ben't one o' th' handsomest, she's noan faal and varry good-natured; and i' his een she's fair beautiful, onybody may
			see that.' I wrote to Moor House and to Cambridge immediately, to say what I had done: fully explaining also why I had thus acted. Diana and Mary approved
			the step unreservedly. Diana announced that she would just give me time to get over the honeymoon, and then she would come and see me.'She had better not
			wait till then, Jane,' said Mr. Rochester, when I read her letter to him; 'if she does, she will be too late, for our honeymoon will shine our life long:
			its beams will only fade over your grave or mine.'How St. John received the news, I don't know: he never answered the letter in which I communicated it:
			yet six months after he wrote to me, without, however, mentioning Mr. Rochester's name or alluding to my marriage. His letter was then calm, and, though
			very serious, kind. He has maintained a regular, though not frequent, correspondence ever since: he hopes I am happy, and trusts I am not of those who
			live without God in the world, and only mind earthly things. You have not quite forgotten little Adele, have you, reader? I had not; I soon asked and
			obtained leave of Mr. Rochester, to go and see her at the school where he had placed her. Her frantic joy at beholding me again moved me much. She looked
			pale and thin: she said she was not happy. I found the rules of the establishment were too strict, its course of study too severe for a child of her
			age: I took her home with me. I meant to become her governess once more, but I soon found this impracticable; my time and cares were now required by
			another- my husband needed them all. So I sought out a school conducted on a more indulgent system, and near enough to permit of my visiting her often,
			and bringing her home sometimes. I took care she should never want for anything that could contribute to her comfort: she soon settled in her new abode,
			became very happy there, and made fair progress in her studies. As she grew up, a sound English education corrected in a great measure her French defects;
			and when she left school, I found in her a pleasing and obliging companion: docile, good-tempered, and well-principled. By her grateful attention to me
			and mine, she has long since well repaid any little kindness I ever had it in my power to offer her.`
	--*/

	msg := host.Pad([]byte(testString))

	msgSize := binary.Size(msg)
	numBlocks := uint32(msgSize >> 6)

	// set up input & output memory buffers
	inputBuff := world.Malloc(xcl.ReadOnly, uint(msgSize))
	defer inputBuff.Free()
	outputBuff := world.Malloc(xcl.WriteOnly, 32)
	defer outputBuff.Free()

	// re-order bytes before write to buffer
	msgOrdered := make([]byte, 0)
	revBytes := make([]byte, 4)
	for i := 0; i < msgSize/4; i++ {
		revBytes = reverseBytes(msg[i*4 : i*4+4])
		msgOrdered = append(msgOrdered, revBytes...)
	}
	binary.Write(inputBuff.Writer(), binary.LittleEndian, msgOrdered)

	// pass args to kernel & start it
	krnl.SetArg(0, numBlocks)
	krnl.SetMemoryArg(1, inputBuff)
	krnl.SetMemoryArg(2, outputBuff)
	krnl.Run(1, 1, 1)

	// reorder bytes after read from buffer
	ret := make([]byte, 32)
	retOrdered := make([]byte, 0)
	err := binary.Read(outputBuff.Reader(), binary.LittleEndian, ret) // outputBuffer -> ret
	for i := 0; i < 8; i++ {
		revBytes = reverseBytes(ret[i*4 : i*4+4])
		retOrdered = append(retOrdered, revBytes...)
	}
	if err != nil {
		log.Fatal("binary.Read failed:", err)
	}

	s := hex.EncodeToString(retOrdered)

	// fox test string
	if s != "ef537f25c895bfa782526529a9b63d97aa631564d5d789c2b765448c8635fb6c" {
		log.Fatalf("%s != %s %v", s, "ef537f25c895bfa782526529a9b63d97aa631564d5d789c2b765448c8635fb6c", ballotX)
	} else {
		log.Printf("PASS: Got hex string of %s %v", s, checkMark)
	}

	//	// JaneEyre test string
	//	if s != "3ad7214e9b9dfd59246332a0bc7f5a17fd428342e75ce7882ef184b999a1bdf0" {
	//		log.Fatalf("%s != %s %v", s, "3ad7214e9b9dfd59246332a0bc7f5a17fd428342e75ce7882ef184b999a1bdf0", ballotX)
	//	} else {
	//		log.Printf("PASS: Got hex string of %s %v", s, checkMark)
	//	}

}
